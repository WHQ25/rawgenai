package image

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
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
		Short:         "Download completed image",
		Long:          "Download the result of a completed image generation task.",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.jpg or .png)")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string, flags *downloadFlags) error {
	taskID := args[0]
	if taskID == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}

	// Validate output extension
	lowerOutput := strings.ToLower(flags.output)
	if !strings.HasSuffix(lowerOutput, ".jpg") && !strings.HasSuffix(lowerOutput, ".jpeg") && !strings.HasSuffix(lowerOutput, ".png") {
		return common.WriteError(cmd, "invalid_output", "output file must have .jpg, .jpeg, or .png extension")
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("LUMA_API_KEY"))
	}

	// Get generation status to get download URL
	req, err := shared.CreateRequest("GET", "/generations/"+taskID, nil)
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

	var gen shared.Generation
	if err := json.NewDecoder(resp.Body).Decode(&gen); err != nil {
		return common.WriteError(cmd, "decode_error", err.Error())
	}

	// Check generation state
	if gen.State != shared.StateCompleted {
		return common.WriteError(cmd, "task_not_ready",
			"task is not ready for download. Current state: "+gen.State)
	}

	// Get image URL
	if gen.Assets == nil || gen.Assets.Image == "" {
		return common.WriteError(cmd, "no_output", "task completed but no image available")
	}
	downloadURL := gen.Assets.Image

	// Download file
	client := &http.Client{Timeout: 2 * time.Minute}
	downloadResp, err := client.Get(downloadURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", "failed to download: "+err.Error())
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != 200 {
		return common.WriteError(cmd, "download_error",
			"download failed with status: "+downloadResp.Status)
	}

	// Create output directory if needed
	dir := filepath.Dir(flags.output)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return common.WriteError(cmd, "output_error", "failed to create directory: "+err.Error())
		}
	}

	// Write file
	outFile, err := os.Create(flags.output)
	if err != nil {
		return common.WriteError(cmd, "output_error", "failed to create output file: "+err.Error())
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, downloadResp.Body); err != nil {
		return common.WriteError(cmd, "output_error", "failed to write output file: "+err.Error())
	}

	// Return absolute path
	absPath, _ := filepath.Abs(flags.output)

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"file":    absPath,
	})
}
