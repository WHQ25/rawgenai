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

var validLanguages = map[string]bool{
	"en": true, "hi": true, "pt": true, "zh": true, "es": true,
	"fr": true, "de": true, "ja": true, "ar": true, "ru": true,
	"ko": true, "id": true, "it": true, "nl": true, "tr": true,
	"pl": true, "sv": true, "fil": true, "ms": true, "ro": true,
	"uk": true, "el": true, "cs": true, "da": true, "fi": true,
	"bg": true, "hr": true, "sk": true, "ta": true,
}

type dubbingFlags struct {
	input        string
	lang         string
	noClone      bool
	noBackground bool
	speakers     int
}

func newDubbingCmd() *cobra.Command {
	flags := &dubbingFlags{}

	cmd := &cobra.Command{
		Use:           "dubbing",
		Short:         "Dub audio to another language",
		Long:          "Dub audio content to a target language.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDubbing(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.input, "input", "i", "", "Input audio file (URL or local path)")
	cmd.Flags().StringVarP(&flags.lang, "lang", "l", "", "Target language code (required)")
	cmd.Flags().BoolVar(&flags.noClone, "no-clone", false, "Disable voice cloning")
	cmd.Flags().BoolVar(&flags.noBackground, "no-background", false, "Remove background audio")
	cmd.Flags().IntVar(&flags.speakers, "speakers", 0, "Number of speakers (auto-detect if 0)")

	return cmd
}

func runDubbing(cmd *cobra.Command, args []string, flags *dubbingFlags) error {
	// 1. Validate required: input
	if flags.input == "" {
		return common.WriteError(cmd, "missing_input", "input file is required (-i)")
	}

	// 2. Validate required: lang
	if flags.lang == "" {
		return common.WriteError(cmd, "missing_lang", "target language is required (-l)")
	}

	// 3. Validate enum: lang
	if !validLanguages[flags.lang] {
		return common.WriteError(cmd, "invalid_lang", "invalid language code")
	}

	// 4. Validate file existence (local files only)
	if !shared.IsURL(flags.input) {
		if _, err := os.Stat(flags.input); os.IsNotExist(err) {
			return common.WriteError(cmd, "input_not_found", "input file not found: "+flags.input)
		}
	}

	// 5. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 6. Resolve input URI
	inputURI, err := shared.ResolveMediaURI(flags.input, "audio")
	if err != nil {
		return common.WriteError(cmd, "input_read_error", "failed to read input: "+err.Error())
	}

	// 7. Build request body
	body := map[string]any{
		"model":                "eleven_voice_dubbing",
		"audioUri":             inputURI,
		"targetLang":           flags.lang,
		"disableVoiceCloning":  flags.noClone,
		"dropBackgroundAudio":  flags.noBackground,
	}
	if flags.speakers > 0 {
		body["numSpeakers"] = flags.speakers
	}

	// 8. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/voice_dubbing", bytes.NewReader(bodyJSON))
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

	// 9. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 10. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
