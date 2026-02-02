package video

import (
	"bytes"
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var (
	validT2VModels = map[string]bool{
		"veo3.1":      true,
		"veo3.1_fast": true,
		"veo3":        true,
	}
	validT2VRatios = map[string]bool{
		"1280:720":  true,
		"720:1280":  true,
		"1080:1920": true,
		"1920:1080": true,
	}
	validT2VDurations = map[int]bool{
		4: true,
		6: true,
		8: true,
	}
)

type text2videoFlags struct {
	model      string
	ratio      string
	duration   int
	audio      bool
	promptFile string
}

func newText2VideoCmd() *cobra.Command {
	flags := &text2videoFlags{}

	cmd := &cobra.Command{
		Use:           "text2video <prompt>",
		Short:         "Generate video from text",
		Long:          "Generate a video from a text prompt using Runway AI models.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runText2Video(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.model, "model", "m", "veo3.1", "Model: veo3.1, veo3.1_fast, veo3")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "1280:720", "Output resolution")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 4, "Duration: 4, 6, or 8 seconds")
	cmd.Flags().BoolVar(&flags.audio, "audio", true, "Generate audio with video")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runText2Video(cmd *cobra.Command, args []string, flags *text2videoFlags) error {
	// 1. Get prompt (required)
	prompt, err := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", "prompt is required")
	}

	// 2. Validate enum: model
	if !validT2VModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "invalid model. Valid models: veo3.1, veo3.1_fast, veo3")
	}

	// 3. Validate enum: ratio
	if !validT2VRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "invalid ratio. Valid ratios: 1280:720, 720:1280, 1080:1920, 1920:1080")
	}

	// 4. Validate enum: duration
	if !validT2VDurations[flags.duration] {
		return common.WriteError(cmd, "invalid_duration", "duration must be 4, 6, or 8 seconds")
	}

	// 5. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 6. Build request body
	body := map[string]any{
		"model":      flags.model,
		"promptText": prompt,
		"ratio":      flags.ratio,
		"duration":   flags.duration,
		"audio":      flags.audio,
	}

	// 7. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/text_to_video", bytes.NewReader(bodyJSON))
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

	// 8. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 9. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
