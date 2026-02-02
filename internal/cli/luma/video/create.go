package video

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

var validVideoModels = map[string]bool{
	"ray-2":       true,
	"ray-flash-2": true,
}

var validAspectRatios = map[string]bool{
	"1:1":  true,
	"16:9": true,
	"9:16": true,
	"4:3":  true,
	"3:4":  true,
	"21:9": true,
	"9:21": true,
}

var validDurations = map[string]bool{
	"5s": true,
	"9s": true,
}

var validResolutions = map[string]bool{
	"540p":  true,
	"720p":  true,
	"1080p": true,
	"4k":    true,
}

type createFlags struct {
	image      string
	endFrame   string
	model      string
	ratio      string
	duration   string
	resolution string
	loop       bool
	promptFile string
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:   "create [prompt]",
		Short: "Create a video from text or image",
		Long:  "Generate a video using Luma Dream Machine. Supports text-to-video and image-to-video.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Start frame image (local file or URL)")
	cmd.Flags().StringVar(&flags.endFrame, "end-frame", "", "End frame image (local file or URL)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "ray-2", "Model (ray-2, ray-flash-2)")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "16:9", "Aspect ratio (1:1, 16:9, 9:16, 4:3, 3:4, 21:9, 9:21)")
	cmd.Flags().StringVarP(&flags.duration, "duration", "d", "5s", "Duration (5s, 9s)")
	cmd.Flags().StringVar(&flags.resolution, "resolution", "", "Resolution (540p, 720p, 1080p, 4k)")
	cmd.Flags().BoolVar(&flags.loop, "loop", false, "Create looping video")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt (optional for image-to-video)
	prompt, _ := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())

	// Validate model
	if !validVideoModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "model must be ray-2 or ray-flash-2")
	}

	// Validate ratio
	if !validAspectRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "ratio must be 1:1, 16:9, 9:16, 4:3, 3:4, 21:9, or 9:21")
	}

	// Validate duration
	if !validDurations[flags.duration] {
		return common.WriteError(cmd, "invalid_duration", "duration must be 5s or 9s")
	}

	// Validate resolution if provided
	if flags.resolution != "" && !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", "resolution must be 540p, 720p, 1080p, or 4k")
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Build keyframes if images provided
	var keyframes map[string]interface{}
	if flags.image != "" || flags.endFrame != "" {
		keyframes = make(map[string]interface{})

		if flags.image != "" {
			// Check if local file exists
			if !shared.IsURL(flags.image) {
				if _, err := os.Stat(flags.image); os.IsNotExist(err) {
					return common.WriteError(cmd, "image_not_found", "start frame image not found: "+flags.image)
				}
			}
			imgURL, err := shared.ResolveImageURL(flags.image)
			if err != nil {
				return common.WriteError(cmd, "image_read_error", err.Error())
			}
			keyframes["frame0"] = map[string]string{
				"type": "image",
				"url":  imgURL,
			}
		}

		if flags.endFrame != "" {
			// Check if local file exists
			if !shared.IsURL(flags.endFrame) {
				if _, err := os.Stat(flags.endFrame); os.IsNotExist(err) {
					return common.WriteError(cmd, "image_not_found", "end frame image not found: "+flags.endFrame)
				}
			}
			imgURL, err := shared.ResolveImageURL(flags.endFrame)
			if err != nil {
				return common.WriteError(cmd, "image_read_error", err.Error())
			}
			keyframes["frame1"] = map[string]string{
				"type": "image",
				"url":  imgURL,
			}
		}
	}

	// Build request body
	body := map[string]interface{}{
		"model":        flags.model,
		"aspect_ratio": flags.ratio,
		"duration":     flags.duration,
	}

	if prompt != "" {
		body["prompt"] = prompt
	}

	if flags.loop {
		body["loop"] = true
	}

	if flags.resolution != "" {
		body["resolution"] = flags.resolution
	}

	if keyframes != nil {
		body["keyframes"] = keyframes
	}

	jsonBody, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/generations/video", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return shared.HandleAPIError(cmd, resp)
	}

	var gen shared.Generation
	if err := json.NewDecoder(resp.Body).Decode(&gen); err != nil {
		return common.WriteError(cmd, "decode_error", err.Error())
	}

	return common.WriteSuccess(cmd, map[string]interface{}{
		"task_id":    gen.ID,
		"state":      gen.State,
		"model":      gen.Model,
		"created_at": gen.CreatedAt,
	})
}
