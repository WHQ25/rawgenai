package elevenlabs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

const baseURL = "https://api.elevenlabs.io/v1"

// Voice name to ID mapping for default voices
var defaultVoices = map[string]string{
	"rachel":  "21m00Tcm4TlvDq8ikWAM",
	"domi":    "AZnzlk1XvdvUeBnXmlld",
	"bella":   "EXAVITQu4vr4xnSDxMaL",
	"antoni":  "ErXwobaYiN019PkySvjV",
	"elli":    "MF3mGyEYCl7XYWbV9V6O",
	"josh":    "TxGEqnHWrfWFTfGW9XjX",
	"arnold":  "VR6AewLTigWG4xSOukaG",
	"adam":    "pNInz6obpgDQGcFmaJgB",
	"sam":     "yoZ06aMxZJJ28mfd3POQ",
	"charlie": "IKne3meq5aSn9XLyUdCD",
}

var ttsOutputFormats = map[string]bool{
	"mp3_22050_32":  true,
	"mp3_44100_128": true,
	"mp3_44100_192": true,
	"pcm_16000":     true,
	"pcm_22050":     true,
	"pcm_24000":     true,
	"pcm_44100":     true,
}

type ttsFlags struct {
	output     string
	promptFile string
	voice      string
	model      string
	format     string
	stability  float64
	similarity float64
	style      float64
	speed      float64
	speak      bool
}

type ttsResponse struct {
	Success    bool   `json:"success"`
	File       string `json:"file,omitempty"`
	Voice      string `json:"voice,omitempty"`
	Model      string `json:"model,omitempty"`
	Characters int    `json:"characters,omitempty"`
}

type ttsRequestBody struct {
	Text          string        `json:"text"`
	ModelID       string        `json:"model_id"`
	VoiceSettings voiceSettings `json:"voice_settings"`
}

type voiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
	Style           float64 `json:"style"`
	Speed           float64 `json:"speed"`
}

var ttsCmd = newTTSCmd()

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:           "tts [text]",
		Short:         "Text to Speech using ElevenLabs voices",
		Long:          "Convert text to speech using ElevenLabs TTS models with high-quality voices.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp3, .wav, .pcm)")
	cmd.Flags().StringVar(&flags.promptFile, "file", "", "Input text file")
	cmd.Flags().StringVarP(&flags.voice, "voice", "v", "Rachel", "Voice name or ID")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "eleven_multilingual_v2", "Model ID")
	cmd.Flags().StringVarP(&flags.format, "format", "f", "mp3_44100_128", "Output format")
	cmd.Flags().Float64Var(&flags.stability, "stability", 0.5, "Voice stability (0.0-1.0)")
	cmd.Flags().Float64Var(&flags.similarity, "similarity", 0.75, "Similarity boost (0.0-1.0)")
	cmd.Flags().Float64Var(&flags.style, "style", 0.0, "Style exaggeration (0.0-1.0)")
	cmd.Flags().Float64Var(&flags.speed, "speed", 1.0, "Speaking speed (0.25-4.0)")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	return cmd
}

func runTTS(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	// Get text from args, file, or stdin
	text, err := getText(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	// Validate output or speak
	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	// Determine output path and format
	var outputPath string
	var useTempFile bool
	outputFormat := flags.format

	if flags.output != "" {
		outputPath = flags.output
	} else {
		// --speak only: use temp file with mp3 format
		outputFormat = "mp3_44100_128"
		tmpFile, err := os.CreateTemp("", "tts-*.mp3")
		if err != nil {
			return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create temp file: %s", err.Error()))
		}
		outputPath = tmpFile.Name()
		tmpFile.Close()
		useTempFile = true
	}

	// Validate format
	if !ttsOutputFormats[outputFormat] {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("unsupported format '%s', supported: mp3_22050_32, mp3_44100_128, mp3_44100_192, pcm_16000, pcm_22050, pcm_24000, pcm_44100", outputFormat))
	}

	// Validate speed
	if flags.speed < 0.25 || flags.speed > 4.0 {
		return common.WriteError(cmd, "invalid_speed", "speed must be between 0.25 and 4.0")
	}

	// Validate stability
	if flags.stability < 0.0 || flags.stability > 1.0 {
		return common.WriteError(cmd, "invalid_stability", "stability must be between 0.0 and 1.0")
	}

	// Validate similarity
	if flags.similarity < 0.0 || flags.similarity > 1.0 {
		return common.WriteError(cmd, "invalid_similarity", "similarity must be between 0.0 and 1.0")
	}

	// Validate style
	if flags.style < 0.0 || flags.style > 1.0 {
		return common.WriteError(cmd, "invalid_style", "style must be between 0.0 and 1.0")
	}

	// Check API key
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "ELEVENLABS_API_KEY environment variable is not set")
	}

	// Resolve voice ID
	voiceID := resolveVoiceID(flags.voice)

	// Build request body
	reqBody := ttsRequestBody{
		Text:    text,
		ModelID: flags.model,
		VoiceSettings: voiceSettings{
			Stability:       flags.stability,
			SimilarityBoost: flags.similarity,
			Style:           flags.style,
			Speed:           flags.speed,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	// Make API request
	url := fmt.Sprintf("%s/text-to-speech/%s?output_format=%s", baseURL, voiceID, outputFormat)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return handleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	// Handle API errors
	if resp.StatusCode != http.StatusOK {
		if useTempFile {
			os.Remove(outputPath)
		}
		return handleAPIErrorResponse(cmd, resp)
	}

	// Get absolute path for output
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	// Write to file
	outFile, err := os.Create(absPath)
	if err != nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	// Play audio if --speak is set
	if flags.speak {
		outFile.Close() // Close before playing
		if err := common.PlayFile(absPath); err != nil {
			if useTempFile {
				os.Remove(absPath)
			}
			return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", err.Error()))
		}
		if useTempFile {
			os.Remove(absPath)
		}
	}

	// Return success
	result := ttsResponse{
		Success:    true,
		File:       absPath,
		Voice:      flags.voice,
		Model:      flags.model,
		Characters: len(text),
	}
	if useTempFile {
		result.File = "" // Don't report temp file path
	}
	return common.WriteSuccess(cmd, result)
}

func resolveVoiceID(voice string) string {
	// Check if it's a known voice name (case-insensitive)
	if id, ok := defaultVoices[strings.ToLower(voice)]; ok {
		return id
	}
	// Assume it's already a voice ID
	return voice
}

func getText(args []string, filePath string, stdin io.Reader) (string, error) {
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
			return "", fmt.Errorf("cannot read file: %w", err)
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	// From stdin (only if not a terminal)
	if stdin != nil {
		// Check if stdin is a terminal (skip if it is)
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				// Is a terminal, skip
				return "", errors.New("no text provided, use positional argument, --file flag, or pipe from stdin")
			}
		}
		// Read from stdin (either piped file or other io.Reader like in tests)
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %w", err)
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			return text, nil
		}
	}

	return "", errors.New("no text provided, use positional argument, --file flag, or pipe from stdin")
}

type apiErrorResponse struct {
	Detail struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"detail"`
}

func handleAPIErrorResponse(cmd *cobra.Command, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Detail.Message != "" {
		status := apiErr.Detail.Status
		message := apiErr.Detail.Message

		// Match exact ElevenLabs error codes first
		switch status {
		case "quota_exceeded":
			return common.WriteError(cmd, "quota_exceeded", message)
		case "max_character_limit_exceeded":
			return common.WriteError(cmd, "max_character_limit_exceeded", message)
		case "invalid_api_key":
			return common.WriteError(cmd, "invalid_api_key", message)
		case "voice_not_found":
			return common.WriteError(cmd, "voice_not_found", message)
		case "only_for_creator+":
			return common.WriteError(cmd, "subscription_required", message)
		case "too_many_concurrent_requests":
			return common.WriteError(cmd, "too_many_concurrent_requests", message)
		case "system_busy":
			return common.WriteError(cmd, "system_busy", message)
		}

		// Fallback: check message content for quota
		if strings.Contains(message, "quota") {
			return common.WriteError(cmd, "quota_exceeded", message)
		}

		// HTTP status code based fallback
		switch resp.StatusCode {
		case 400:
			return common.WriteError(cmd, "invalid_request", message)
		case 401:
			return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
		case 403:
			return common.WriteError(cmd, "forbidden", message)
		case 404:
			return common.WriteError(cmd, "voice_not_found", message)
		case 422:
			return common.WriteError(cmd, "validation_error", message)
		case 429:
			return common.WriteError(cmd, "rate_limit", message)
		case 500:
			return common.WriteError(cmd, "server_error", "ElevenLabs server error")
		case 503:
			return common.WriteError(cmd, "server_overloaded", "ElevenLabs server overloaded")
		default:
			return common.WriteError(cmd, "api_error", message)
		}
	}

	// Fallback for non-JSON or unparseable errors
	switch resp.StatusCode {
	case 401:
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
	case 404:
		return common.WriteError(cmd, "voice_not_found", "Voice not found")
	case 429:
		return common.WriteError(cmd, "rate_limit", "Too many requests")
	default:
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API error: %d", resp.StatusCode))
	}
}

func handleHTTPError(cmd *cobra.Command, err error) error {
	if strings.Contains(err.Error(), "timeout") {
		return common.WriteError(cmd, "timeout", "Request timed out")
	}
	return common.WriteError(cmd, "connection_error", fmt.Sprintf("Cannot connect to ElevenLabs API: %s", err.Error()))
}
