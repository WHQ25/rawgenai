package video

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type downloadFlags struct {
	output string
}

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <task_id>",
		Short:         "Download completed video",
		Long:          "Download the result of a completed video generation task.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp4)")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string, flags *downloadFlags) error {
	// 1. Validate required: task_id
	if len(args) == 0 {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}
	taskID := args[0]

	// 2. Validate required: output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}

	// 3. Validate output extension
	if !strings.HasSuffix(strings.ToLower(flags.output), ".mp4") {
		return common.WriteError(cmd, "invalid_output", "output file must have .mp4 extension")
	}

	// 4. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 5. Get task status to get download URL
	req, err := shared.CreateRequest("GET", "/v1/tasks/"+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return shared.HandleAPIError(cmd, resp)
	}

	var taskStatus shared.TaskStatus
	if err := json.NewDecoder(resp.Body).Decode(&taskStatus); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 6. Check task status
	if taskStatus.Status != shared.StatusSucceeded {
		return common.WriteError(cmd, "task_not_ready",
			"task is not ready for download. Current status: "+taskStatus.Status)
	}

	// 7. Get download URL
	if len(taskStatus.Output) == 0 {
		return common.WriteError(cmd, "no_output", "task completed but no output available")
	}
	downloadURL := taskStatus.Output[0]

	// 8. Download file
	client := &http.Client{Timeout: 5 * time.Minute}
	downloadResp, err := client.Get(downloadURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", "failed to download: "+err.Error())
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != 200 {
		return common.WriteError(cmd, "download_error",
			"download failed with status: "+downloadResp.Status)
	}

	// 9. Create output directory if needed
	dir := filepath.Dir(flags.output)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return common.WriteError(cmd, "output_error", "failed to create directory: "+err.Error())
		}
	}

	// 10. Write file
	outFile, err := os.Create(flags.output)
	if err != nil {
		return common.WriteError(cmd, "output_error", "failed to create output file: "+err.Error())
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, downloadResp.Body); err != nil {
		return common.WriteError(cmd, "output_error", "failed to write output file: "+err.Error())
	}

	// 11. Return absolute path
	absPath, _ := filepath.Abs(flags.output)

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"file":    absPath,
	})
}
