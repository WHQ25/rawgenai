package video

import (
	"bytes"
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

type upscaleFlags struct {
	resolution string
}

func newUpscaleCmd() *cobra.Command {
	flags := &upscaleFlags{}

	cmd := &cobra.Command{
		Use:   "upscale <task_id>",
		Short: "Upscale a video generation",
		Long:  "Upscale an existing video generation to higher resolution.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpscale(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVar(&flags.resolution, "resolution", "1080p", "Target resolution (540p, 720p, 1080p, 4k)")

	return cmd
}

func runUpscale(cmd *cobra.Command, args []string, flags *upscaleFlags) error {
	taskID := args[0]
	if taskID == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	// Validate resolution
	if !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", "resolution must be 540p, 720p, 1080p, or 4k")
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Build request body
	body := map[string]interface{}{
		"generation_type": "upscale_video",
		"resolution":      flags.resolution,
	}

	jsonBody, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/generations/"+taskID+"/upscale", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return shared.HandleAPIError(cmd, resp)
	}

	var gen shared.Generation
	if err := json.NewDecoder(resp.Body).Decode(&gen); err != nil {
		return common.WriteError(cmd, "decode_error", err.Error())
	}

	return common.WriteSuccess(cmd, map[string]interface{}{
		"task_id":    gen.ID,
		"state":      gen.State,
		"created_at": gen.CreatedAt,
	})
}
