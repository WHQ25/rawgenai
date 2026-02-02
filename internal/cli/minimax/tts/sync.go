package tts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var validFormats = map[string]bool{
	"mp3":  true,
	"pcm":  true,
	"flac": true,
	"wav":  true,
}

var validSampleRates = map[int]bool{
	8000:  true,
	16000: true,
	22050: true,
	24000: true,
	32000: true,
	44100: true,
}

var validBitrates = map[int]bool{
	32000:  true,
	64000:  true,
	128000: true,
	256000: true,
}

type ttsSyncResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file,omitempty"`
	Model   string `json:"model,omitempty"`
	Voice   string `json:"voice,omitempty"`
}

func runSync(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	text, err := getText(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	if err := validateSyncFlags(cmd, flags); err != nil {
		return err
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	if flags.stream {
		return runWebsocket(cmd, text, flags)
	}

	body := map[string]any{
		"model":         flags.model,
		"text":          text,
		"stream":        false,
		"output_format": "hex",
		"voice_setting": map[string]any{
			"voice_id": flags.voice,
			"speed":    flags.speed,
			"vol":      flags.vol,
			"pitch":    flags.pitch,
		},
		"audio_setting": map[string]any{
			"format": flags.format,
		},
	}

	if flags.sampleRate != 0 {
		body["audio_setting"].(map[string]any)["sample_rate"] = flags.sampleRate
	}
	if flags.bitrate != 0 {
		body["audio_setting"].(map[string]any)["bitrate"] = flags.bitrate
	}
	if flags.channel != 0 {
		body["audio_setting"].(map[string]any)["channel"] = flags.channel
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/t2a_v2", bytes.NewReader(jsonBody))
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
		Data struct {
			Audio  string `json:"audio"`
			Status int    `json:"status"`
		} `json:"data"`
		BaseResp struct {
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

	if apiResp.Data.Audio == "" {
		return common.WriteError(cmd, "empty_audio", "no audio data returned")
	}

	audioBytes, err := hex.DecodeString(apiResp.Data.Audio)
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

	return common.WriteSuccess(cmd, ttsSyncResponse{
		Success: true,
		File:    absPath,
		Model:   flags.model,
		Voice:   flags.voice,
	})
}
