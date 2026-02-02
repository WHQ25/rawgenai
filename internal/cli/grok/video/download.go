package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type downloadFlags struct {
	output string
}

type downloadResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"request_id"`
	File      string `json:"file"`
}

// API response for status check
type xaiVideoDownloadStatusResponse struct {
	RequestID string         `json:"request_id"`
	Status    string         `json:"status"`
	VideoURL  string         `json:"video_url,omitempty"`
	Error     *xaiVideoError `json:"error,omitempty"`
}

var downloadCmd = newDownloadCmd()

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <request_id>",
		Short:         "Download generated video",
		Long:          "Download video from a completed generation or edit operation.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp4)")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string, flags *downloadFlags) error {
	requestID := strings.TrimSpace(args[0])
	if requestID == "" {
		return common.WriteError(cmd, "missing_request_id", "request_id is required")
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	ext := strings.ToLower(filepath.Ext(flags.output))
	if ext != ".mp4" {
		return common.WriteError(cmd, "invalid_format", "output file must be .mp4")
	}

	// Check API key
	apiKey := config.GetAPIKey("XAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("XAI_API_KEY"))
	}

	// First, check status to get video URL
	statusURL := fmt.Sprintf("%s%s/%s", xaiBaseURL, videosPath, requestID)
	statusReq, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	statusReq.Header.Set("Authorization", "Bearer "+apiKey)

	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		return handleHTTPError(cmd, err)
	}
	defer statusResp.Body.Close()

	// Parse status response
	var apiResp xaiVideoDownloadStatusResponse
	if err := json.NewDecoder(statusResp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for API error
	if apiResp.Error != nil {
		return handleAPIError(cmd, statusResp.StatusCode, apiResp.Error)
	}

	if statusResp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d", statusResp.StatusCode))
	}

	// Check status
	switch apiResp.Status {
	case "completed", "succeeded":
		// Ready to download
	case "failed":
		errMsg := "video generation failed"
		if apiResp.Error != nil {
			errMsg = apiResp.Error.Message
		}
		return common.WriteError(cmd, "video_failed", errMsg)
	default:
		return common.WriteError(cmd, "video_not_ready", fmt.Sprintf("video is not ready for download, current status: %s", apiResp.Status))
	}

	// Check video URL
	if apiResp.VideoURL == "" {
		return common.WriteError(cmd, "no_video", "video URL not available")
	}

	// Download video
	videoResp, err := http.Get(apiResp.VideoURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download video: %s", err.Error()))
	}
	defer videoResp.Body.Close()

	if videoResp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("video download failed with status %d", videoResp.StatusCode))
	}

	// Get absolute path
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	// Write to file
	outFile, err := os.Create(absPath)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, videoResp.Body)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	return common.WriteSuccess(cmd, downloadResponse{
		Success:   true,
		RequestID: requestID,
		File:      absPath,
	})
}
