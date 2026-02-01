package video

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

type downloadFlags struct {
	output string
}

type downloadResponse struct {
	Success     bool   `json:"success"`
	OperationID string `json:"operation_id"`
	File        string `json:"file"`
}

var downloadCmd = newDownloadCmd()

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <operation_id>",
		Short:         "Download generated video",
		Long:          "Download video from a completed generation operation.",
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
	operationID := strings.TrimSpace(args[0])
	if operationID == "" {
		return common.WriteError(cmd, "missing_operation_id", "operation_id is required")
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
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "GEMINI_API_KEY or GOOGLE_API_KEY environment variable is not set")
	}

	// Create client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return common.WriteError(cmd, "client_error", err.Error())
	}

	// Create operation reference with just the name
	opRef := &genai.GenerateVideosOperation{
		Name: operationID,
	}

	// Get operation status
	op, err := client.Operations.GetVideosOperation(ctx, opRef, nil)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Check if video is ready
	if !op.Done {
		return common.WriteError(cmd, "video_not_ready", "Video is not ready for download, current status: running")
	}

	// Check if there was an error
	if op.Error != nil {
		if msg, ok := op.Error["message"].(string); ok && msg != "" {
			return common.WriteError(cmd, "video_failed", fmt.Sprintf("Video generation failed: %s", msg))
		}
	}

	// Get video URL from response
	if op.Response == nil || len(op.Response.GeneratedVideos) == 0 {
		return common.WriteError(cmd, "no_video", "no video generated in response")
	}

	video := op.Response.GeneratedVideos[0].Video
	if video == nil {
		return common.WriteError(cmd, "no_video", "video data not available")
	}

	// Get absolute path for output
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	// Download video content using SDK (handles authentication)
	if len(video.VideoBytes) == 0 && video.URI != "" {
		_, err = client.Files.Download(ctx, video, nil)
		if err != nil {
			return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download video: %s", err.Error()))
		}
	}

	// Write video bytes to file
	if len(video.VideoBytes) == 0 {
		return common.WriteError(cmd, "no_video", "video data not available after download")
	}

	if err := os.WriteFile(absPath, video.VideoBytes, 0644); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	result := downloadResponse{
		Success:     true,
		OperationID: operationID,
		File:        absPath,
	}
	return common.WriteSuccess(cmd, result)
}
