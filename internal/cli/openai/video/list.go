package video

import (
	"context"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	oai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/spf13/cobra"
)

type listFlags struct {
	limit int
	order string
}

type listItem struct {
	VideoID   string `json:"video_id"`
	Status    string `json:"status"`
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	Size      string `json:"size"`
	Duration  string `json:"duration"`
	Progress  int64  `json:"progress"`
	CreatedAt int64  `json:"created_at"`
}

type listResponse struct {
	Success bool       `json:"success"`
	Videos  []listItem `json:"videos"`
	Count   int        `json:"count"`
}

var listCmd = newListCmd()

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List all video generation jobs",
		Long:          "List all video generation jobs with their status.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flags)
		},
	}

	cmd.Flags().IntVarP(&flags.limit, "limit", "n", 20, "Maximum number of videos to return")
	cmd.Flags().StringVar(&flags.order, "order", "desc", "Sort order by timestamp (asc, desc)")

	return cmd
}

func runList(cmd *cobra.Command, flags *listFlags) error {
	// Validate order
	if flags.order != "asc" && flags.order != "desc" {
		return common.WriteError(cmd, "invalid_order", "order must be 'asc' or 'desc'")
	}

	// Check API key
	apiKey := config.GetAPIKey("OPENAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("OPENAI_API_KEY"))
	}

	client := oai.NewClient(option.WithAPIKey(apiKey))
	ctx := context.Background()

	params := oai.VideoListParams{
		Limit: oai.Int(int64(flags.limit)),
		Order: oai.VideoListParamsOrder(flags.order),
	}

	page, err := client.Videos.List(ctx, params)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	videos := make([]listItem, 0)
	for _, v := range page.Data {
		videos = append(videos, listItem{
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

	result := listResponse{
		Success: true,
		Videos:  videos,
		Count:   len(videos),
	}

	return common.WriteSuccess(cmd, result)
}
