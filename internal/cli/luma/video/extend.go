package video

import (
	"bytes"
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

type extendFlags struct {
	reverse    bool
	model      string
	ratio      string
	promptFile string
}

func newExtendCmd() *cobra.Command {
	flags := &extendFlags{}

	cmd := &cobra.Command{
		Use:   "extend <task_id> [prompt]",
		Short: "Extend a video generation",
		Long:  "Extend an existing video by using it as the start (or end with --reverse) frame.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtend(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().BoolVar(&flags.reverse, "reverse", false, "Use generation as end frame (prepend to video)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "ray-2", "Model (ray-2, ray-flash-2)")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "16:9", "Aspect ratio")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runExtend(cmd *cobra.Command, args []string, flags *extendFlags) error {
	taskID := args[0]
	if taskID == "" {
		return common.WriteError(cmd, "missing_task_id", "task_id is required")
	}

	// Get prompt (optional)
	promptArgs := args[1:]
	prompt, _ := shared.GetPrompt(promptArgs, flags.promptFile, nil)

	// Validate model
	if !validVideoModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "model must be ray-2 or ray-flash-2")
	}

	// Validate ratio
	if !validAspectRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "ratio must be 1:1, 16:9, 9:16, 4:3, 3:4, 21:9, or 9:21")
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Build keyframes - reference the existing generation
	keyframes := make(map[string]interface{})
	frameKey := "frame0" // Default: use generation as start frame
	if flags.reverse {
		frameKey = "frame1" // Use generation as end frame (prepend)
	}
	keyframes[frameKey] = map[string]string{
		"type": "generation",
		"id":   taskID,
	}

	// Build request body
	body := map[string]interface{}{
		"model":        flags.model,
		"aspect_ratio": flags.ratio,
		"keyframes":    keyframes,
	}

	if prompt != "" {
		body["prompt"] = prompt
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
