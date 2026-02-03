package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/kling/video"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type ttsFlags struct {
	output     string
	promptFile string
	voiceID    string
	language   string
	speed      float64
	speak      bool
}

var Cmd = newTTSCmd()

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:           "tts [text]",
		Short:         "Kling text-to-speech",
		Long:          "Convert text to speech using Kling TTS.",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read text from file")
	cmd.Flags().StringVar(&flags.voiceID, "voice", "", "Voice ID (required)")
	cmd.Flags().StringVar(&flags.language, "language", "zh", "Voice language: zh, en")
	cmd.Flags().Float64Var(&flags.speed, "speed", 1.0, "Speech speed (0.8-2.0)")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	return cmd
}

func runTTS(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	// Get text
	text, err := video.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	if strings.TrimSpace(flags.voiceID) == "" {
		return common.WriteError(cmd, "missing_voice", "--voice is required")
	}

	if flags.language != "zh" && flags.language != "en" {
		return common.WriteError(cmd, "invalid_language", "language must be zh or en")
	}

	if flags.speed < 0.8 || flags.speed > 2.0 {
		return common.WriteError(cmd, "invalid_speed", "speed must be between 0.8 and 2.0")
	}

	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"text":           text,
		"voice_id":       flags.voiceID,
		"voice_language": flags.language,
		"voice_speed":    flags.speed,
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", video.GetKlingAPIBase()+"/v1/audio/tts", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return video.HandleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			TaskID        string `json:"task_id"`
			TaskStatus    string `json:"task_status"`
			TaskStatusMsg string `json:"task_status_msg"`
			TaskResult    *struct {
				Audios []struct {
					ID       string `json:"id"`
					URL      string `json:"url"`
					Duration string `json:"duration"`
				} `json:"audios"`
			} `json:"task_result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	if result.Data.TaskStatus == "failed" {
		msg := result.Data.TaskStatusMsg
		if msg == "" {
			msg = "tts failed"
		}
		return common.WriteError(cmd, "tts_failed", msg)
	}

	if result.Data.TaskStatus != "succeed" {
		return common.WriteError(cmd, "tts_not_ready", "task is not finished")
	}

	if result.Data.TaskResult == nil || len(result.Data.TaskResult.Audios) == 0 {
		return common.WriteError(cmd, "response_error", "no audio in response")
	}

	audio := result.Data.TaskResult.Audios[0]
	if strings.TrimSpace(audio.URL) == "" {
		return common.WriteError(cmd, "response_error", "audio URL is empty")
	}

	outputPath := flags.output
	useTempFile := false

	if outputPath == "" {
		ext := getURLExt(audio.URL)
		if ext == "" {
			ext = ".mp3"
		}
		tmpFile, err := os.CreateTemp("", "kling-tts-*"+ext)
		if err != nil {
			return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create temp file: %s", err.Error()))
		}
		outputPath = tmpFile.Name()
		tmpFile.Close()
		useTempFile = true
	}

	if flags.speak {
		ext := strings.ToLower(filepath.Ext(outputPath))
		if ext != ".mp3" && ext != ".wav" {
			if useTempFile {
				os.Remove(outputPath)
			}
			return common.WriteError(cmd, "unsupported_format", "--speak only supports mp3 or wav")
		}
	}

	if !useTempFile {
		// Create output directory if needed
		dir := filepath.Dir(outputPath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create directory: %s", err.Error()))
			}
		}
	}

	// Download audio
	downloadClient := &http.Client{Timeout: 5 * time.Minute}
	downloadResp, err := downloadClient.Get(audio.URL)
	if err != nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download audio: %s", err.Error()))
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status: %d", downloadResp.StatusCode))
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create file: %s", err.Error()))
	}

	if _, err := io.Copy(outFile, downloadResp.Body); err != nil {
		outFile.Close()
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot write file: %s", err.Error()))
	}
	outFile.Close()

	if flags.speak {
		if err := common.PlayFile(outputPath); err != nil {
			if useTempFile {
				os.Remove(outputPath)
			}
			return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", err.Error()))
		}
		if useTempFile {
			os.Remove(outputPath)
		}
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	resultPayload := map[string]any{
		"success":  true,
		"task_id":  result.Data.TaskID,
		"status":   result.Data.TaskStatus,
		"voice_id": flags.voiceID,
		"duration": audio.Duration,
	}

	if !useTempFile {
		resultPayload["file"] = absPath
	}

	return common.WriteSuccess(cmd, resultPayload)
}

func getURLExt(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(path.Ext(parsed.Path))
}
