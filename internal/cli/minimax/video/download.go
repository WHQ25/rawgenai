package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type downloadFlags struct {
	output string
}

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <file_id>",
		Short:         "Download generated video",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args[0], flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp4)")

	return cmd
}

func runDownload(cmd *cobra.Command, fileID string, flags *downloadFlags) error {
	if fileID == "" {
		return common.WriteError(cmd, "missing_file_id", "file_id is required")
	}
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}
	if !strings.HasSuffix(strings.ToLower(flags.output), ".mp4") {
		return common.WriteError(cmd, "invalid_output", "output file must have .mp4 extension")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	req, err := shared.CreateRequest("GET", "/v1/files/retrieve?file_id="+fileID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var apiResp struct {
		File struct {
			DownloadURL string `json:"download_url"`
		} `json:"file"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg))
	}
	if apiResp.File.DownloadURL == "" {
		return common.WriteError(cmd, "download_error", "download_url is empty")
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	downloadResp, err := client.Get(apiResp.File.DownloadURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download file: %s", err.Error()))
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status: %s", downloadResp.Status))
	}

	dir := filepath.Dir(flags.output)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return common.WriteError(cmd, "output_error", fmt.Sprintf("failed to create directory: %s", err.Error()))
		}
	}

	outFile, err := os.Create(flags.output)
	if err != nil {
		return common.WriteError(cmd, "output_error", fmt.Sprintf("failed to create output file: %s", err.Error()))
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, downloadResp.Body); err != nil {
		return common.WriteError(cmd, "output_error", fmt.Sprintf("failed to write output file: %s", err.Error()))
	}

	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"file_id": fileID,
		"file":    absPath,
	})
}
