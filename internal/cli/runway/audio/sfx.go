package audio

import (
	"bytes"
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type sfxFlags struct {
	duration   float64
	loop       bool
	promptFile string
}

func newSfxCmd() *cobra.Command {
	flags := &sfxFlags{}

	cmd := &cobra.Command{
		Use:           "sfx <prompt>",
		Short:         "Generate sound effects",
		Long:          "Generate sound effects from a text description.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSfx(cmd, args, flags)
		},
	}

	cmd.Flags().Float64VarP(&flags.duration, "duration", "d", 0, "Duration in seconds (0.5-30, auto if not set)")
	cmd.Flags().BoolVar(&flags.loop, "loop", false, "Create seamless loop")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runSfx(cmd *cobra.Command, args []string, flags *sfxFlags) error {
	// 1. Get prompt (required)
	prompt, err := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", "prompt is required")
	}

	// 2. Validate range: duration
	if flags.duration != 0 && (flags.duration < 0.5 || flags.duration > 30) {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 0.5 and 30 seconds")
	}

	// 3. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 4. Build request body
	body := map[string]any{
		"model":      "eleven_text_to_sound_v2",
		"promptText": prompt,
		"loop":       flags.loop,
	}
	if flags.duration > 0 {
		body["duration"] = flags.duration
	}

	// 5. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/sound_effect", bytes.NewReader(bodyJSON))
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

	// 6. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 7. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
