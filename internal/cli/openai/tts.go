package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var supportedFormats = map[string]oai.AudioSpeechNewParamsResponseFormat{
	".mp3":  oai.AudioSpeechNewParamsResponseFormatMP3,
	".opus": oai.AudioSpeechNewParamsResponseFormatOpus,
	".aac":  oai.AudioSpeechNewParamsResponseFormatAAC,
	".flac": oai.AudioSpeechNewParamsResponseFormatFLAC,
	".wav":  oai.AudioSpeechNewParamsResponseFormatWAV,
	".pcm":  oai.AudioSpeechNewParamsResponseFormatPCM,
}

type ttsFlags struct {
	output       string
	promptFile   string
	voice        string
	model        string
	instructions string
	speed        float64
}

type ttsResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file,omitempty"`
	Model   string `json:"model,omitempty"`
	Voice   string `json:"voice,omitempty"`
}

var ttsCmd = newTTSCmd()

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:           "tts [text]",
		Short:         "Text to Speech using OpenAI TTS models",
		Long:          "Convert text to speech using OpenAI TTS models (gpt-4o-mini-tts, tts-1, tts-1-hd).",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (format from extension)")
	cmd.Flags().StringVar(&flags.promptFile, "file", "", "Input text file")
	cmd.Flags().StringVar(&flags.voice, "voice", "alloy", "Voice name or custom voice ID")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "gpt-4o-mini-tts", "Model name")
	cmd.Flags().StringVar(&flags.instructions, "instructions", "", "Voice style instructions (gpt-4o-mini-tts only)")
	cmd.Flags().Float64Var(&flags.speed, "speed", 1, "Speed (0.25 - 4.0)")

	return cmd
}

func runTTS(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	// Get text from args, file, or stdin
	text, err := getText(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate format
	ext := strings.ToLower(filepath.Ext(flags.output))
	responseFormat, ok := supportedFormats[ext]
	if !ok {
		return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s', supported: mp3, opus, aac, flac, wav, pcm", ext))
	}

	// Validate speed
	if flags.speed < 0.25 || flags.speed > 4.0 {
		return common.WriteError(cmd, "invalid_speed", "speed must be between 0.25 and 4.0")
	}

	// Validate instructions compatibility
	if flags.instructions != "" && (flags.model == "tts-1" || flags.model == "tts-1-hd") {
		return common.WriteError(cmd, "invalid_parameter", "--instructions is not supported by model '"+flags.model+"', use 'gpt-4o-mini-tts' instead")
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	// Call OpenAI API
	client := oai.NewClient()
	ctx := context.Background()

	params := oai.AudioSpeechNewParams{
		Model:          oai.SpeechModel(flags.model),
		Voice:          oai.AudioSpeechNewParamsVoice(flags.voice),
		Input:          text,
		ResponseFormat: responseFormat,
		Speed:          oai.Float(flags.speed),
	}

	// Add instructions if provided (only for gpt-4o-mini-tts)
	if flags.instructions != "" {
		params.Instructions = oai.String(flags.instructions)
	}

	resp, err := client.Audio.Speech.New(ctx, params)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Get absolute path for output
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	// Write to file
	outFile, err := os.Create(absPath)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	// Return success
	result := ttsResponse{
		Success: true,
		File:    absPath,
		Model:   flags.model,
		Voice:   flags.voice,
	}
	return common.WriteSuccess(cmd, result)
}

func handleAPIError(cmd *cobra.Command, err error) error {
	var apiErr *oai.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 400:
			return common.WriteError(cmd, "invalid_request", apiErr.Message)
		case 401:
			return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
		case 403:
			return common.WriteError(cmd, "region_not_supported", "Region/country not supported")
		case 429:
			if strings.Contains(apiErr.Message, "quota") {
				return common.WriteError(cmd, "quota_exceeded", apiErr.Message)
			}
			return common.WriteError(cmd, "rate_limit", apiErr.Message)
		case 500:
			return common.WriteError(cmd, "server_error", "OpenAI server error")
		case 503:
			return common.WriteError(cmd, "server_overloaded", "OpenAI server overloaded")
		default:
			return common.WriteError(cmd, "api_error", apiErr.Message)
		}
	}

	// Network errors
	if strings.Contains(err.Error(), "timeout") {
		return common.WriteError(cmd, "timeout", "Request timed out")
	}
	return common.WriteError(cmd, "connection_error", fmt.Sprintf("Cannot connect to OpenAI API: %s", err.Error()))
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
