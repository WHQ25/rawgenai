package openai

import (
	"context"
	"os"

	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

type videoListFlags struct {
	limit int
	order string
}

type videoListItem struct {
	VideoID   string `json:"video_id"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	Size      string `json:"size"`
	Duration  string `json:"duration"`
	Progress  int64  `json:"progress"`
	CreatedAt int64  `json:"created_at"`
}

type videoListResponse struct {
	Success bool            `json:"success"`
	Videos  []videoListItem `json:"videos"`
	Count   int             `json:"count"`
}

var videoListCmd = newVideoListCmd()

func newVideoListCmd() *cobra.Command {
	flags := &videoListFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List all video generation jobs",
		Long:          "List all video generation jobs with their status.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoList(cmd, flags)
		},
	}

	cmd.Flags().IntVarP(&flags.limit, "limit", "n", 20, "Maximum number of videos to return")
	cmd.Flags().StringVar(&flags.order, "order", "desc", "Sort order by timestamp (asc, desc)")

	return cmd
}

func runVideoList(cmd *cobra.Command, flags *videoListFlags) error {
	// Validate order
	if flags.order != "asc" && flags.order != "desc" {
		return writeError(cmd, "invalid_order", "order must be 'asc' or 'desc'")
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return writeError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	client := oai.NewClient()
	ctx := context.Background()

	params := oai.VideoListParams{
		Limit: oai.Int(int64(flags.limit)),
		Order: oai.VideoListParamsOrder(flags.order),
	}

	page, err := client.Videos.List(ctx, params)
	if err != nil {
		return handleVideoAPIError(cmd, err)
	}

	videos := make([]videoListItem, 0)
	for _, v := range page.Data {
		videos = append(videos, videoListItem{
			VideoID:   v.ID,
			Status:    string(v.Status),
			Model:     string(v.Model),
			Prompt:    v.Prompt,
			Size:      string(v.Size),
			Duration:  string(v.Seconds),
			Progress:  v.Progress,
			CreatedAt: v.CreatedAt,
		})
	}

	result := videoListResponse{
		Success: true,
		Videos:  videos,
		Count:   len(videos),
	}

	return writeSuccess(cmd, result)
}
