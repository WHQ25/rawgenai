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

type deleteResponse struct {
	Success bool   `json:"success"`
	VideoID string `json:"video_id"`
	Deleted bool   `json:"deleted"`
}

var deleteCmd = newDeleteCmd()

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete <video_id>",
		Short:         "Delete a video",
		Long:          "Delete a video and its associated assets.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args)
		},
	}

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	videoID := strings.TrimSpace(args[0])
	if videoID == "" {
		return common.WriteError(cmd, "missing_video_id", "video_id is required")
	}

	// Check API key
	apiKey := config.GetAPIKey("OPENAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("OPENAI_API_KEY"))
	}

	client := oai.NewClient(option.WithAPIKey(apiKey))
	ctx := context.Background()

	resp, err := client.Videos.Delete(ctx, videoID)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	result := deleteResponse{
		Success: true,
		VideoID: resp.ID,
		Deleted: resp.Deleted,
	}

	return common.WriteSuccess(cmd, result)
}
