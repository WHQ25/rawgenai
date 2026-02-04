package dashscope

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

const (
	ttsGenerationPath = "/services/aigc/multimodal-generation/generation"
)

// Valid TTS models
var (
	validHTTPTTSModels = map[string]bool{
		"qwen3-tts-flash":            true,
		"qwen3-tts-flash-2025-11-27": true,
		"qwen3-tts-flash-2025-09-18": true,
		"qwen-tts":                   true,
		"qwen-tts-latest":            true,
	}

	validRealtimeTTSModels = map[string]bool{
		"qwen3-tts-flash-realtime":                   true,
		"qwen3-tts-flash-realtime-2025-11-27":        true,
		"qwen3-tts-instruct-flash-realtime":          true,
		"qwen3-tts-instruct-flash-realtime-2026-01-22": true,
	}

	validLanguages = map[string]bool{
		"Auto":       true,
		"Chinese":    true,
		"English":    true,
		"Japanese":   true,
		"Korean":     true,
		"French":     true,
		"German":     true,
		"Russian":    true,
		"Italian":    true,
		"Spanish":    true,
		"Portuguese": true,
	}

	validSampleRates = map[int]bool{
		24000: true,
		48000: true,
	}

	// HTTP models only support WAV
	httpTTSFormats = map[string]bool{
		".wav": true,
	}

	// Realtime models support multiple formats
	realtimeTTSFormats = map[string]string{
		".mp3":  "mp3",
		".pcm":  "pcm",
		".opus": "opus",
		".wav":  "wav",
	}
)

type ttsFlags struct {
	output       string
	promptFile   string
	voice        string
	model        string
	language     string
	instructions string
	sampleRate   int
	speak        bool
}

type ttsSuccessResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file,omitempty"`
	Model   string `json:"model"`
	Voice   string `json:"voice"`
}

var ttsCmd = newTTSCmd()

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:           "tts [text]",
		Short:         "Text to Speech using Qwen-TTS models",
		Long:          "Convert text to speech using Alibaba Qwen-TTS models via DashScope API.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (format from extension)")
	cmd.Flags().StringVarP(&flags.promptFile, "file", "f", "", "Input text file")
	cmd.Flags().StringVar(&flags.voice, "voice", "Cherry", "Voice name")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "qwen3-tts-flash", "Model name")
	cmd.Flags().StringVarP(&flags.language, "language", "l", "Auto", "Language type")
	cmd.Flags().StringVar(&flags.instructions, "instructions", "", "Style instructions (instruct model only)")
	cmd.Flags().IntVar(&flags.sampleRate, "sample-rate", 24000, "Sample rate in Hz (realtime only)")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	return cmd
}

func isRealtimeModel(model string) bool {
	return strings.Contains(model, "-realtime")
}

func isInstructModel(model string) bool {
	return strings.Contains(model, "-instruct-")
}

// countCharacters counts characters using DashScope rules:
// 1 Chinese character = 2 characters, 1 ASCII char = 1 character
func countCharacters(text string) int {
	count := 0
	for _, r := range text {
		if utf8.RuneLen(r) > 1 {
			count += 2
		} else {
			count++
		}
	}
	return count
}

func runTTS(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	// Get text from args, file, or stdin
	text, err := getTTSText(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	// Validate output or speak
	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	// Validate model
	realtime := isRealtimeModel(flags.model)
	if !validHTTPTTSModels[flags.model] && !validRealtimeTTSModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s'", flags.model))
	}

	// Determine output path and format
	var outputPath string
	var useTempFile bool
	var ext string

	if flags.output != "" {
		outputPath = flags.output
		ext = strings.ToLower(filepath.Ext(outputPath))
	} else {
		// --speak only: use temp file
		if realtime {
			ext = ".mp3"
		} else {
			ext = ".wav"
		}
		tmpFile, tmpErr := os.CreateTemp("", "tts-*"+ext)
		if tmpErr != nil {
			return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create temp file: %s", tmpErr.Error()))
		}
		outputPath = tmpFile.Name()
		tmpFile.Close()
		useTempFile = true
	}

	// Validate format
	if realtime {
		if _, ok := realtimeTTSFormats[ext]; !ok {
			return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s' for realtime model, supported: mp3, pcm, opus, wav", ext))
		}
	} else {
		if !httpTTSFormats[ext] {
			return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s' for HTTP model, only .wav is supported", ext))
		}
	}

	// Validate language
	if !validLanguages[flags.language] {
		return common.WriteError(cmd, "invalid_language", fmt.Sprintf("invalid language '%s', supported: Auto, Chinese, English, Japanese, Korean, French, German, Russian, Italian, Spanish, Portuguese", flags.language))
	}

	// Validate instructions compatibility
	if flags.instructions != "" && !isInstructModel(flags.model) {
		return common.WriteError(cmd, "incompatible_instructions", fmt.Sprintf("--instructions is only supported by instruct models, not '%s'", flags.model))
	}

	// Validate sample rate
	if cmd.Flags().Changed("sample-rate") {
		if !realtime {
			return common.WriteError(cmd, "incompatible_sample_rate", "--sample-rate is only supported by realtime models")
		}
		if !validSampleRates[flags.sampleRate] {
			return common.WriteError(cmd, "invalid_sample_rate", fmt.Sprintf("invalid sample rate %d, supported: 24000, 48000", flags.sampleRate))
		}
	}

	// Validate text length for HTTP models
	if !realtime && countCharacters(text) > 600 {
		return common.WriteError(cmd, "text_too_long", fmt.Sprintf("text exceeds 600 characters (got %d), use a realtime model for longer text", countCharacters(text)))
	}

	// Check API key
	apiKey := config.GetAPIKey("DASHSCOPE_API_KEY")
	if apiKey == "" {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("DASHSCOPE_API_KEY"))
	}

	// Get absolute path for output
	absPath, pathErr := filepath.Abs(outputPath)
	if pathErr != nil {
		absPath = outputPath
	}

	// Call appropriate API
	if realtime {
		err = runTTSRealtime(cmd, text, absPath, ext, apiKey, flags)
	} else {
		err = runTTSHTTP(cmd, text, absPath, apiKey, flags)
	}

	if err != nil {
		if useTempFile {
			os.Remove(absPath)
		}
		return err
	}

	// Play audio if --speak is set
	if flags.speak {
		if playErr := common.PlayFile(absPath); playErr != nil {
			if useTempFile {
				os.Remove(absPath)
			}
			return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", playErr.Error()))
		}
		if useTempFile {
			os.Remove(absPath)
		}
	}

	// Return success
	result := ttsSuccessResponse{
		Success: true,
		File:    absPath,
		Model:   flags.model,
		Voice:   flags.voice,
	}
	if useTempFile {
		result.File = ""
	}
	return common.WriteSuccess(cmd, result)
}

// runTTSHTTP calls the synchronous HTTP API for non-realtime models
func runTTSHTTP(cmd *cobra.Command, text, outputPath, apiKey string, flags *ttsFlags) error {
	baseURL := getBaseURL()

	// Build request body
	reqBody := map[string]any{
		"model": flags.model,
		"input": map[string]any{
			"text":          text,
			"voice":         flags.voice,
			"language_type": flags.language,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	req, err := http.NewRequest("POST", baseURL+ttsGenerationPath, bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return handleHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	// Parse response
	var result struct {
		Output *struct {
			Audio *struct {
				URL  string `json:"url"`
				Data string `json:"data"`
			} `json:"audio"`
		} `json:"output"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != "" {
		if strings.Contains(result.Code, "DataInspection") || strings.Contains(result.Code, "Infringement") {
			return common.WriteError(cmd, "content_policy", result.Message)
		}
		return common.WriteError(cmd, "invalid_request", result.Message)
	}

	if result.Output == nil || result.Output.Audio == nil {
		return common.WriteError(cmd, "response_error", "no audio in response")
	}

	// Download audio from URL
	if result.Output.Audio.URL != "" {
		return downloadAudioURL(cmd, result.Output.Audio.URL, outputPath)
	}

	// Or decode base64 audio data
	if result.Output.Audio.Data != "" {
		audioData, decErr := base64.StdEncoding.DecodeString(result.Output.Audio.Data)
		if decErr != nil {
			return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot decode audio data: %s", decErr.Error()))
		}
		return os.WriteFile(outputPath, audioData, 0644)
	}

	return common.WriteError(cmd, "response_error", "no audio URL or data in response")
}

// runTTSRealtime connects via WebSocket for realtime TTS models
func runTTSRealtime(cmd *cobra.Command, text, outputPath, ext, apiKey string, flags *ttsFlags) error {
	baseURL := getBaseURL()
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	// Replace /api/v1 with /api-ws/v1/realtime
	wsURL = strings.Replace(wsURL, "/api/v1", "/api-ws/v1/realtime", 1)
	wsURL += "?model=" + flags.model

	// Connect WebSocket
	header := http.Header{}
	header.Set("Authorization", "Bearer "+apiKey)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot connect to WebSocket: %s", err.Error()))
	}
	defer conn.Close()

	// Send session.update
	sessionUpdate := map[string]any{
		"type": "session.update",
		"session": map[string]any{
			"mode":            "commit",
			"voice":           flags.voice,
			"language_type":   flags.language,
			"response_format": realtimeTTSFormats[ext],
			"sample_rate":     flags.sampleRate,
		},
	}
	if flags.instructions != "" {
		session := sessionUpdate["session"].(map[string]any)
		session["instructions"] = flags.instructions
	}

	if err := conn.WriteJSON(sessionUpdate); err != nil {
		return common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot send session update: %s", err.Error()))
	}

	// Wait for session.updated
	if err := waitForEvent(conn, "session.updated"); err != nil {
		return common.WriteError(cmd, "websocket_error", fmt.Sprintf("session update failed: %s", err.Error()))
	}

	// Send text
	appendMsg := map[string]any{
		"type": "input_text_buffer.append",
		"text": text,
	}
	if err := conn.WriteJSON(appendMsg); err != nil {
		return common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot send text: %s", err.Error()))
	}

	// Commit then signal no more input
	commitMsg := map[string]any{
		"type": "input_text_buffer.commit",
	}
	if err := conn.WriteJSON(commitMsg); err != nil {
		return common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot commit: %s", err.Error()))
	}

	finishMsg := map[string]any{"type": "session.finish"}
	if err := conn.WriteJSON(finishMsg); err != nil {
		return common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot send finish: %s", err.Error()))
	}

	// Collect audio chunks until session ends
	outFile, err := os.Create(outputPath)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	for {
		_, message, readErr := conn.ReadMessage()
		if readErr != nil {
			// Connection closed after session.finished is normal
			break
		}

		var event struct {
			Type    string `json:"type"`
			Delta   string `json:"delta"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		switch event.Type {
		case "response.audio.delta":
			audioData, decErr := base64.StdEncoding.DecodeString(event.Delta)
			if decErr != nil {
				return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot decode audio chunk: %s", decErr.Error()))
			}
			if _, writeErr := outFile.Write(audioData); writeErr != nil {
				return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write audio: %s", writeErr.Error()))
			}
		case "session.finished":
			_ = outFile.Sync()
			return nil
		case "error":
			return common.WriteError(cmd, "server_error", fmt.Sprintf("server error: %s", event.Message))
		}
	}

	_ = outFile.Sync()
	return nil
}

func downloadAudioURL(cmd *cobra.Command, audioURL, outputPath string) error {
	resp, err := http.Get(audioURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download audio: %s", err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status %d", resp.StatusCode))
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	return nil
}

func waitForEvent(conn *websocket.Conn, expectedType string) error {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("connection closed: %s", err.Error())
		}

		var event struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}
		if jsonErr := json.Unmarshal(message, &event); jsonErr != nil {
			continue
		}

		if event.Type == expectedType {
			return nil
		}
		if event.Type == "error" {
			return fmt.Errorf("%s", event.Message)
		}
	}
}

func getTTSText(args []string, filePath string, stdin io.Reader) (string, error) {
	// From positional argument
	if len(args) > 0 {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text != "" {
			return text, nil
		}
	}

	// From file
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	// From stdin
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", fmt.Errorf("no text provided, use positional argument, --file flag, or pipe from stdin")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	return "", fmt.Errorf("no text provided, use positional argument, --file flag, or pipe from stdin")
}
