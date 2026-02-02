package voice

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type designFlags struct {
	prompt      string
	promptFile  string
	previewText string
	previewFile string
	voiceID     string
	output      string
	speak       bool
}

func newDesignCmd() *cobra.Command {
	flags := &designFlags{}

	cmd := &cobra.Command{
		Use:           "design",
		Short:         "Generate a new voice from text description",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDesign(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.prompt, "prompt", "", "Voice description prompt")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read prompt from file")
	cmd.Flags().StringVar(&flags.previewText, "preview-text", "", "Preview text for trial audio")
	cmd.Flags().StringVar(&flags.previewFile, "preview-file", "", "Read preview text from file")
	cmd.Flags().StringVar(&flags.voiceID, "voice-id", "", "Optional voice ID to reuse")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path for trial audio")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play trial audio after generation (mp3)")

	return cmd
}

func runDesign(cmd *cobra.Command, flags *designFlags) error {
	prompt, err := resolveText(flags.prompt, flags.promptFile)
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}
	previewText, err := resolveText(flags.previewText, flags.previewFile)
	if err != nil {
		return common.WriteError(cmd, "missing_preview_text", err.Error())
	}

	if prompt == "" {
		return common.WriteError(cmd, "missing_prompt", "prompt is required")
	}
	if previewText == "" {
		return common.WriteError(cmd, "missing_preview_text", "preview-text is required")
	}

	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := map[string]any{
		"prompt":       prompt,
		"preview_text": previewText,
	}
	if strings.TrimSpace(flags.voiceID) != "" {
		body["voice_id"] = flags.voiceID
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/voice_design", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var apiResp struct {
		TrialAudio string `json:"trial_audio"`
		VoiceID    string `json:"voice_id"`
		BaseResp   struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg))
	}

	if apiResp.TrialAudio == "" {
		return common.WriteError(cmd, "empty_audio", "trial audio is empty")
	}

	audioBytes, err := hex.DecodeString(apiResp.TrialAudio)
	if err != nil {
		return common.WriteError(cmd, "decode_error", fmt.Sprintf("cannot decode audio: %s", err.Error()))
	}

	var absPath string
	if flags.output != "" {
		absPath, err = filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}
		if err := os.WriteFile(absPath, audioBytes, 0644); err != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
		}
	}

	if flags.speak {
		reader := bytes.NewReader(audioBytes)
		if err := common.PlayMP3(reader); err != nil {
			return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", err.Error()))
		}
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success":  true,
		"voice_id": apiResp.VoiceID,
		"file":     absPath,
	})
}

func resolveText(text string, filePath string) (string, error) {
	if strings.TrimSpace(text) != "" {
		return strings.TrimSpace(text), nil
	}
	if filePath == "" {
		return "", nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %s", err.Error())
	}
	return strings.TrimSpace(string(data)), nil
}
