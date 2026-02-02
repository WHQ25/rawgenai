package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type cloneFlags struct {
	fileID              int64
	voiceID             string
	promptAudioFileID   int64
	promptText          string
	previewText         string
	model               string
	languageBoost       string
	noiseReduction      bool
	volumeNormalization bool
	continuousSound     bool
}

func newCloneCmd() *cobra.Command {
	flags := &cloneFlags{}

	cmd := &cobra.Command{
		Use:           "clone",
		Short:         "Clone a voice from uploaded audio",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(cmd, flags)
		},
	}

	cmd.Flags().Int64Var(&flags.fileID, "file-id", 0, "Uploaded audio file ID (required)")
	cmd.Flags().StringVar(&flags.voiceID, "voice-id", "", "Voice ID to create (required)")
	cmd.Flags().Int64Var(&flags.promptAudioFileID, "prompt-audio-id", 0, "Prompt audio file ID (optional)")
	cmd.Flags().StringVar(&flags.promptText, "prompt-text", "", "Prompt transcript (required if prompt-audio-id is set)")
	cmd.Flags().StringVar(&flags.previewText, "preview-text", "", "Preview text to generate demo audio")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "speech-2.8-hd", "Model name for preview")
	cmd.Flags().StringVar(&flags.languageBoost, "language-boost", "", "Language boost (optional)")
	cmd.Flags().BoolVar(&flags.noiseReduction, "noise-reduction", false, "Enable noise reduction")
	cmd.Flags().BoolVar(&flags.volumeNormalization, "volume-normalization", false, "Enable volume normalization")
	cmd.Flags().BoolVar(&flags.continuousSound, "continuous-sound", false, "Enable continuous sound")

	return cmd
}

func runClone(cmd *cobra.Command, flags *cloneFlags) error {
	if flags.fileID == 0 {
		return common.WriteError(cmd, "missing_file_id", "file-id is required")
	}
	if strings.TrimSpace(flags.voiceID) == "" {
		return common.WriteError(cmd, "missing_voice_id", "voice-id is required")
	}
	if (flags.promptAudioFileID != 0 && strings.TrimSpace(flags.promptText) == "") ||
		(flags.promptAudioFileID == 0 && strings.TrimSpace(flags.promptText) != "") {
		return common.WriteError(cmd, "invalid_prompt", "prompt-audio-id and prompt-text must be provided together")
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := map[string]any{
		"file_id":  flags.fileID,
		"voice_id": flags.voiceID,
	}

	if flags.promptAudioFileID != 0 {
		body["clone_prompt"] = map[string]any{
			"prompt_audio": flags.promptAudioFileID,
			"prompt_text":  flags.promptText,
		}
	}

	if strings.TrimSpace(flags.previewText) != "" {
		body["text"] = flags.previewText
		if strings.TrimSpace(flags.model) != "" {
			body["model"] = flags.model
		}
	}

	if flags.languageBoost != "" {
		body["language_boost"] = flags.languageBoost
	}
	if flags.noiseReduction {
		body["need_noise_reduction"] = true
	}
	if flags.volumeNormalization {
		body["need_volume_normalization"] = true
	}
	if flags.continuousSound {
		body["continuous_sound"] = true
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/voice_clone", bytes.NewReader(jsonBody))
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
		InputSensitive bool   `json:"input_sensitive"`
		DemoAudio      string `json:"demo_audio"`
		BaseResp       struct {
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

	return common.WriteSuccess(cmd, map[string]any{
		"success":         true,
		"voice_id":        flags.voiceID,
		"demo_audio":      apiResp.DemoAudio,
		"input_sensitive": apiResp.InputSensitive,
	})
}
