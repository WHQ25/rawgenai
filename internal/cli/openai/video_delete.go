package openai

import (
	"context"
	"os"
	"strings"

	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

type videoDeleteResponse struct {
	Success bool   `json:"success"`
	VideoID string `json:"video_id"`
	Deleted bool   `json:"deleted"`
}

var videoDeleteCmd = newVideoDeleteCmd()

func newVideoDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete <video_id>",
		Short:         "Delete a video",
		Long:          "Delete a video and its associated assets.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoDelete(cmd, args)
		},
	}

	return cmd
}

func runVideoDelete(cmd *cobra.Command, args []string) error {
	videoID := strings.TrimSpace(args[0])
	if videoID == "" {
		return writeError(cmd, "missing_video_id", "video_id is required")
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return writeError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	client := oai.NewClient()
	ctx := context.Background()

	resp, err := client.Videos.Delete(ctx, videoID)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}

	result := videoDeleteResponse{
		Success: true,
		VideoID: resp.ID,
		Deleted: resp.Deleted,
	}

	return writeSuccess(cmd, result)
}
