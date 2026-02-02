package video

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

var validModifyModes = map[string]bool{
	"adhere_1":    true,
	"adhere_2":    true,
	"adhere_3":    true,
	"flex_1":      true,
	"flex_2":      true,
	"flex_3":      true,
	"reimagine_1": true,
	"reimagine_2": true,
	"reimagine_3": true,
}

type modifyFlags struct {
	video      string
	mode       string
	model      string
	firstFrame string
	promptFile string
}

func newModifyCmd() *cobra.Command {
	flags := &modifyFlags{}

	cmd := &cobra.Command{
		Use:   "modify [prompt]",
		Short: "Modify a video with style transfer",
		Long: `Modify a video with style transfer and prompt-based editing.

Modes:
  adhere_1, adhere_2, adhere_3  - Stick closely to original motion
  flex_1, flex_2, flex_3        - Balance between original and new motion
  reimagine_1, reimagine_2, reimagine_3 - More creative freedom`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModify(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVarP(&flags.video, "video", "v", "", "Source video (URL required)")
	cmd.Flags().StringVar(&flags.mode, "mode", "", "Modification mode (adhere_1-3, flex_1-3, reimagine_1-3)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "ray-2", "Model (ray-2, ray-flash-2)")
	cmd.Flags().StringVar(&flags.firstFrame, "first-frame", "", "First frame image (URL)")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	cmd.MarkFlagRequired("video")
	cmd.MarkFlagRequired("mode")

	return cmd
}

func runModify(cmd *cobra.Command, args []string, flags *modifyFlags) error {
	// Validate video
	if flags.video == "" {
		return common.WriteError(cmd, "missing_video", "video is required (--video)")
	}

	// Validate mode
	if flags.mode == "" {
		return common.WriteError(cmd, "missing_mode", "mode is required (--mode)")
	}

	if !validModifyModes[flags.mode] {
		return common.WriteError(cmd, "invalid_mode", "mode must be adhere_1-3, flex_1-3, or reimagine_1-3")
	}

	// Validate model
	if !validVideoModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "model must be ray-2 or ray-flash-2")
	}

	// Validate first frame if provided
	if flags.firstFrame != "" && !shared.IsURL(flags.firstFrame) {
		if _, err := os.Stat(flags.firstFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "first frame image not found: "+flags.firstFrame)
		}
	}

	// Get prompt (optional)
	prompt, _ := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Build request body
	body := map[string]interface{}{
		"generation_type": "modify_video",
		"model":           flags.model,
		"mode":            flags.mode,
		"media": map[string]string{
			"url": flags.video,
		},
	}

	if prompt != "" {
		body["prompt"] = prompt
	}

	if flags.firstFrame != "" {
		frameURL, err := shared.ResolveImageURL(flags.firstFrame)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", err.Error())
		}
		body["first_frame"] = map[string]string{
			"url": frameURL,
		}
	}

	jsonBody, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/generations/video/modify", bytes.NewReader(jsonBody))
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
