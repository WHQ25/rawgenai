package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	oai "github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
)

var supportedAudioFormats = map[string]bool{
	".mp3":  true,
	".mp4":  true,
	".mpeg": true,
	".mpga": true,
	".m4a":  true,
	".wav":  true,
	".webm": true,
	".ogg":  true,
	".oga":  true,
	".opus": true,
	".flac": true,
}

var sttResponseFormats = map[string]oai.AudioResponseFormat{
	"json":         oai.AudioResponseFormatJSON,
	"text":         oai.AudioResponseFormatText,
	"srt":          oai.AudioResponseFormatSRT,
	"vtt":          oai.AudioResponseFormatVTT,
	"verbose_json": oai.AudioResponseFormatVerboseJSON,
}

const maxFileSize = 25 * 1024 * 1024 // 25 MB

var errFileNotFound = errors.New("file_not_found")
var errMissingFile = errors.New("missing_file")

type sttFlags struct {
	file        string
	model       string
	language    string
	prompt      string
	temperature float64
	verbose     bool
	format      string
	output      string
}

type sttResponse struct {
	Success  bool         `json:"success"`
	Text     string       `json:"text,omitempty"`
	Model    string       `json:"model,omitempty"`
	Language string       `json:"language,omitempty"`
	Duration float64      `json:"duration,omitempty"`
	Segments []sttSegment `json:"segments,omitempty"`
	File     string       `json:"file,omitempty"`
}

type sttSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

var sttCmd = newSTTCmd()

func newSTTCmd() *cobra.Command {
	flags := &sttFlags{}

	cmd := &cobra.Command{
		Use:           "stt [audio-file]",
		Short:         "Speech to Text using OpenAI Whisper models",
		Long:          "Transcribe audio files to text using OpenAI Whisper models (whisper-1, gpt-4o-transcribe, gpt-4o-mini-transcribe).",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSTT(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "Input audio file path")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "whisper-1", "Model name")
	cmd.Flags().StringVarP(&flags.language, "language", "l", "", "Language code (ISO-639-1)")
	cmd.Flags().StringVar(&flags.prompt, "prompt", "", "Text to guide the model's style or terminology")
	cmd.Flags().Float64Var(&flags.temperature, "temperature", 0, "Sampling temperature (0-1)")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Include timestamps and segments")
	cmd.Flags().StringVar(&flags.format, "format", "json", "Output format (json, text, srt, vtt)")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file (required for srt/vtt formats)")

	return cmd
}

func runSTT(cmd *cobra.Command, args []string, flags *sttFlags) error {
	// Get audio file from args, --file flag, or stdin
	audioFile, audioReader, cleanup, err := getAudioInput(args, flags.file, cmd.InOrStdin())
	if err != nil {
		if errors.Is(err, errFileNotFound) {
			return writeError(cmd, "file_not_found", err.Error())
		}
		return writeError(cmd, "missing_file", err.Error())
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Validate audio format (if we have a file path)
	if audioFile != "" {
		ext := strings.ToLower(filepath.Ext(audioFile))
		if !supportedAudioFormats[ext] {
			return writeError(cmd, "unsupported_format", fmt.Sprintf("unsupported audio format '%s', supported: mp3, mp4, mpeg, mpga, m4a, wav, webm, ogg, oga, opus, flac", ext))
		}

		// Check file size
		info, err := os.Stat(audioFile)
		if err != nil {
			return writeError(cmd, "file_not_found", fmt.Sprintf("cannot access file: %s", err.Error()))
		}
		if info.Size() > maxFileSize {
			return writeError(cmd, "file_too_large", fmt.Sprintf("file size %d bytes exceeds 25 MB limit", info.Size()))
		}
	}

	// Validate temperature
	if flags.temperature < 0 || flags.temperature > 1 {
		return writeError(cmd, "invalid_temperature", "temperature must be between 0 and 1")
	}

	// Validate output format
	responseFormat, ok := sttResponseFormats[flags.format]
	if !ok {
		return writeError(cmd, "invalid_format", fmt.Sprintf("invalid format '%s', supported: json, text, srt, vtt", flags.format))
	}

	// srt/vtt formats require output file
	if (flags.format == "srt" || flags.format == "vtt") && flags.output == "" {
		return writeError(cmd, "missing_output", fmt.Sprintf("--%s format requires --output flag", flags.format))
	}

	// Validate prompt compatibility
	if flags.prompt != "" && flags.model == "gpt-4o-transcribe-diarize" {
		return writeError(cmd, "invalid_parameter", "--prompt is not supported by model 'gpt-4o-transcribe-diarize'")
	}

	// If verbose, use verbose_json format
	if flags.verbose && flags.format == "json" {
		responseFormat = oai.AudioResponseFormatVerboseJSON
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return writeError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	// Open file for API
	var fileReader io.Reader
	var fileToClose *os.File
	if audioReader != nil {
		fileReader = audioReader
	} else {
		f, err := os.Open(audioFile)
		if err != nil {
			return writeError(cmd, "file_not_found", fmt.Sprintf("cannot open file: %s", err.Error()))
		}
		fileToClose = f
		defer f.Close()
		fileReader = f
	}

	// Call OpenAI API
	client := oai.NewClient()
	ctx := context.Background()

	params := oai.AudioTranscriptionNewParams{
		File:           fileReader,
		Model:          oai.AudioModel(flags.model),
		ResponseFormat: responseFormat,
		Temperature:    oai.Float(flags.temperature),
	}

	if flags.language != "" {
		params.Language = oai.String(flags.language)
	}

	if flags.prompt != "" {
		params.Prompt = oai.String(flags.prompt)
	}

	// Avoid unused variable warning
	_ = fileToClose

	resp, err := client.Audio.Transcriptions.New(ctx, params)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Handle output based on format
	if flags.format == "srt" || flags.format == "vtt" {
		// Write subtitle content to file
		absPath, err := filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}

		if err := os.WriteFile(absPath, []byte(resp.Text), 0644); err != nil {
			return writeError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
		}

		result := sttResponse{
			Success:  true,
			File:     absPath,
			Model:    flags.model,
			Language: resp.Language,
		}
		return writeSuccess(cmd, result)
	}

	// Build response
	result := sttResponse{
		Success:  true,
		Text:     resp.Text,
		Model:    flags.model,
		Language: resp.Language,
	}

	// Add verbose info if available
	if flags.verbose || flags.format == "verbose_json" {
		result.Duration = resp.Duration
		if len(resp.Segments) > 0 {
			result.Segments = make([]sttSegment, len(resp.Segments))
			for i, seg := range resp.Segments {
				result.Segments[i] = sttSegment{
					Start: seg.Start,
					End:   seg.End,
					Text:  seg.Text,
				}
			}
		}
	}

	// Write to output file if specified (for json/text formats)
	if flags.output != "" {
		absPath, err := filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}

		if err := os.WriteFile(absPath, []byte(resp.Text), 0644); err != nil {
			return writeError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
		}

		result.File = absPath
	}

	return writeSuccess(cmd, result)
}

func getAudioInput(args []string, filePath string, stdin io.Reader) (file string, reader io.Reader, cleanup func(), err error) {
	// From positional argument
	if len(args) > 0 && args[0] != "" {
		file := strings.TrimSpace(args[0])
		if file != "" {
			if _, err := os.Stat(file); err != nil {
				return "", nil, nil, fmt.Errorf("%w: audio file '%s' does not exist", errFileNotFound, file)
			}
			return file, nil, nil, nil
		}
	}

	// From --file flag
	if filePath != "" {
		if _, err := os.Stat(filePath); err != nil {
			return "", nil, nil, fmt.Errorf("%w: audio file '%s' does not exist", errFileNotFound, filePath)
		}
		return filePath, nil, nil, nil
	}

	// From stdin (only if not a terminal)
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				// Is a terminal, skip
				return "", nil, nil, fmt.Errorf("%w: no audio file provided, use positional argument, --file flag, or pipe from stdin", errMissingFile)
			}
		}

		// Read stdin into temp file (API requires file)
		tmpFile, err := os.CreateTemp("", "stt_stdin_*.mp3")
		if err != nil {
			return "", nil, nil, fmt.Errorf("cannot create temp file: %w", err)
		}

		if _, err := io.Copy(tmpFile, stdin); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", nil, nil, fmt.Errorf("cannot read stdin: %w", err)
		}

		tmpFile.Close()

		cleanup := func() {
			os.Remove(tmpFile.Name())
		}

		return tmpFile.Name(), nil, cleanup, nil
	}

	return "", nil, nil, fmt.Errorf("%w: no audio file provided, use positional argument, --file flag, or pipe from stdin", errMissingFile)
}
