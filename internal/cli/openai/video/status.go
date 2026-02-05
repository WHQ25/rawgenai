package video

import (
	"context"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	oai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/spf13/cobra"
)

type statusResponse struct {
	Success   bool   `json:"success"`
	VideoID   string `json:"video_id"`
	Status    string `json:"status"`
	Error     string `json:"error_message,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

var statusCmd = newStatusCmd()

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <video_id>",
		Short:         "Get video generation status",
		Long:          "Query the status of a video generation job.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args)
		},
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	videoID := strings.TrimSpace(args[0])
	if videoID == "" {
		return common.WriteError(cmd, "missing_video_id", "video_id is required")
	}

	// Check API key
	apiKey := config.GetAPIKey("OPENAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("OPENAI_API_KEY"))
	}

	// Get video status
	client := oai.NewClient(option.WithAPIKey(apiKey))
	ctx := context.Background()

	video, err := client.Videos.Get(ctx, videoID)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	result := statusResponse{
		Success:   true,
		VideoID:   video.ID,
		Status:    string(video.Status),
		CreatedAt: video.CreatedAt,
	}

	// Include error message if failed
	if video.Status == oai.VideoStatusFailed && video.Error.Message != "" {
		result.Error = video.Error.Message
	}

	return common.WriteSuccess(cmd, result)
}
