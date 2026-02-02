package video

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type upscaleFlags struct {
	video string
}

func newUpscaleCmd() *cobra.Command {
	flags := &upscaleFlags{}

	cmd := &cobra.Command{
		Use:           "upscale",
		Short:         "Upscale a video (4x)",
		Long:          "Upscale a video by 4x, capped at 4096px per side.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpscale(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.video, "video", "v", "", "Video to upscale (URL or local path)")

	return cmd
}

func runUpscale(cmd *cobra.Command, args []string, flags *upscaleFlags) error {
	// 1. Validate required: video
	if flags.video == "" {
		return common.WriteError(cmd, "missing_video", "input video is required (-v)")
	}

	// 2. Validate file existence (local files only)
	if !shared.IsURL(flags.video) {
		if _, err := os.Stat(flags.video); os.IsNotExist(err) {
			return common.WriteError(cmd, "video_not_found", "video file not found: "+flags.video)
		}
	}

	// 3. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 4. Resolve video URI
	videoURI, err := shared.ResolveMediaURI(flags.video, "video")
	if err != nil {
		return common.WriteError(cmd, "video_read_error", "failed to read video: "+err.Error())
	}

	// 5. Build request body
	body := map[string]any{
		"model":    "upscale_v1",
		"videoUri": videoURI,
	}

	// 6. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/video_upscale", bytes.NewReader(bodyJSON))
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

	// 7. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 8. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
