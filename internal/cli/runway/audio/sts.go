package audio

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type stsFlags struct {
	input       string
	inputType   string
	voice       string
	removeNoise bool
}

func newSTSCmd() *cobra.Command {
	flags := &stsFlags{}

	cmd := &cobra.Command{
		Use:           "sts",
		Short:         "Convert speech to another voice",
		Long:          "Convert speech from one voice to another in audio or video.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.input, "input", "i", "", "Input audio/video file (URL or local path)")
	cmd.Flags().StringVar(&flags.inputType, "input-type", "audio", "Input type: audio, video")
	cmd.Flags().StringVarP(&flags.voice, "voice", "v", "", "Voice preset ID (required)")
	cmd.Flags().BoolVar(&flags.removeNoise, "remove-noise", false, "Remove background noise")

	return cmd
}

func runSTS(cmd *cobra.Command, args []string, flags *stsFlags) error {
	// 1. Validate required: input
	if flags.input == "" {
		return common.WriteError(cmd, "missing_input", "input file is required (-i)")
	}

	// 2. Validate required: voice
	if flags.voice == "" {
		return common.WriteError(cmd, "missing_voice", "voice preset is required (-v)")
	}

	// 3. Validate enum: inputType
	if flags.inputType != "audio" && flags.inputType != "video" {
		return common.WriteError(cmd, "invalid_input_type", "input-type must be 'audio' or 'video'")
	}

	// 4. Validate enum: voice
	if !validVoices[flags.voice] {
		return common.WriteError(cmd, "invalid_voice", "invalid voice preset")
	}

	// 5. Validate file existence (local files only)
	if !shared.IsURL(flags.input) {
		if _, err := os.Stat(flags.input); os.IsNotExist(err) {
			return common.WriteError(cmd, "input_not_found", "input file not found: "+flags.input)
		}
	}

	// 6. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 7. Resolve input URI
	var inputURI string
	var err error
	if flags.inputType == "audio" {
		inputURI, err = shared.ResolveMediaURI(flags.input, "audio")
	} else {
		inputURI, err = shared.ResolveMediaURI(flags.input, "video")
	}
	if err != nil {
		return common.WriteError(cmd, "input_read_error", "failed to read input: "+err.Error())
	}

	// 8. Build request body
	body := map[string]any{
		"model": "eleven_multilingual_sts_v2",
		"media": map[string]any{
			"type": flags.inputType,
			"uri":  inputURI,
		},
		"voice": map[string]any{
			"type":     "runway-preset",
			"presetId": flags.voice,
		},
		"removeBackgroundNoise": flags.removeNoise,
	}

	// 9. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/speech_to_speech", bytes.NewReader(bodyJSON))
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

	// 10. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 11. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
