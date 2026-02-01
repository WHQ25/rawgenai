package openai

import (
	"context"
	"os"
	"strings"

	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

type videoStatusResponse struct {
	Success   bool   `json:"success"`
	VideoID   string `json:"video_id"`
	Status    string `json:"status"`
	Error     string `json:"error_message,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

var videoStatusCmd = newVideoStatusCmd()

func newVideoStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <video_id>",
		Short:         "Get video generation status",
		Long:          "Query the status of a video generation job.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoStatus(cmd, args)
		},
	}

	return cmd
}

func runVideoStatus(cmd *cobra.Command, args []string) error {
	videoID := strings.TrimSpace(args[0])
	if videoID == "" {
		return writeError(cmd, "missing_video_id", "video_id is required")
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return writeError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	// Get video status
	client := oai.NewClient()
	ctx := context.Background()

	video, err := client.Videos.Get(ctx, videoID)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}

	result := videoStatusResponse{
		Success:   true,
		VideoID:   video.ID,
		Status:    string(video.Status),
		CreatedAt: video.CreatedAt,
	}

	// Include error message if failed
	if video.Status == oai.VideoStatusFailed && video.Error.Message != "" {
		result.Error = video.Error.Message
	}

	return writeSuccess(cmd, result)
}
