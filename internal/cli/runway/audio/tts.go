package audio

import (
	"bytes"
	"encoding/json"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var validVoices = map[string]bool{
	"Maya": true, "Arjun": true, "Serene": true, "Bernard": true, "Billy": true,
	"Mark": true, "Clint": true, "Mabel": true, "Chad": true, "Leslie": true,
	"Eleanor": true, "Elias": true, "Elliot": true, "Grungle": true, "Brodie": true,
	"Sandra": true, "Kirk": true, "Kylie": true, "Lara": true, "Lisa": true,
	"Malachi": true, "Marlene": true, "Martin": true, "Miriam": true, "Monster": true,
	"Paula": true, "Pip": true, "Rusty": true, "Ragnar": true, "Xylar": true,
	"Maggie": true, "Jack": true, "Katie": true, "Noah": true, "James": true,
	"Rina": true, "Ella": true, "Mariah": true, "Frank": true, "Claudia": true,
	"Niki": true, "Vincent": true, "Kendrick": true, "Myrna": true, "Tom": true,
	"Wanda": true, "Benjamin": true, "Kiana": true, "Rachel": true,
}

type ttsFlags struct {
	voice      string
	promptFile string
}

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:           "tts <prompt>",
		Short:         "Generate speech from text",
		Long:          "Generate speech from text using Runway AI.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.voice, "voice", "v", "", "Voice preset ID (required)")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runTTS(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	// 1. Get prompt (required)
	prompt, err := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", "prompt is required")
	}

	// 2. Validate required: voice
	if flags.voice == "" {
		return common.WriteError(cmd, "missing_voice", "voice preset is required (-v)")
	}

	// 3. Validate enum: voice
	if !validVoices[flags.voice] {
		return common.WriteError(cmd, "invalid_voice", "invalid voice preset")
	}

	// 4. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 5. Build request body
	body := map[string]any{
		"model":      "eleven_multilingual_v2",
		"promptText": prompt,
		"voice": map[string]any{
			"type":     "runway-preset",
			"presetId": flags.voice,
		},
	}

	// 6. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/text_to_speech", bytes.NewReader(bodyJSON))
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
