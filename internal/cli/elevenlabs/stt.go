package elevenlabs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

const maxSTTFileSize = 3 * 1024 * 1024 * 1024 // 3GB

var supportedSTTFormats = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".m4a":  true,
	".flac": true,
	".ogg":  true,
	".webm": true,
	".mp4":  true,
	".mkv":  true,
	".mov":  true,
	".avi":  true,
}

type sttFlags struct {
	file       string
	model      string
	language   string
	diarize    bool
	speakers   int
	timestamps string
	output     string
}

type sttResponse struct {
	Success  bool        `json:"success"`
	Text     string      `json:"text,omitempty"`
	Language string      `json:"language,omitempty"`
	Duration float64     `json:"duration,omitempty"`
	Words    []sttWord   `json:"words,omitempty"`
	Speakers []sttSpeaker `json:"speakers,omitempty"`
	File     string      `json:"file,omitempty"`
}

type sttWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type sttSpeaker struct {
	Speaker string  `json:"speaker"`
	Text    string  `json:"text"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
}

type sttAPIResponse struct {
	Text            string `json:"text"`
	LanguageCode    string `json:"language_code"`
	AudioDurationS  float64 `json:"audio_duration_s,omitempty"`
	Words           []sttAPIWord `json:"words,omitempty"`
	Utterances      []sttAPIUtterance `json:"utterances,omitempty"`
}

type sttAPIWord struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Type  string  `json:"type"`
}

type sttAPIUtterance struct {
	SpeakerID string  `json:"speaker_id"`
	Text      string  `json:"text"`
	StartS    float64 `json:"start_s"`
	EndS      float64 `json:"end_s"`
}

var sttCmd = newSTTCmd()

func newSTTCmd() *cobra.Command {
	flags := &sttFlags{}

	cmd := &cobra.Command{
		Use:           "stt [audio-file]",
		Short:         "Speech to Text using ElevenLabs Scribe",
		Long:          "Transcribe audio files to text using ElevenLabs Scribe models with optional speaker diarization.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSTT(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "Input audio file path")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "scribe_v1", "Model: scribe_v1, scribe_v2")
	cmd.Flags().StringVarP(&flags.language, "language", "l", "", "ISO-639 language code (auto-detect if empty)")
	cmd.Flags().BoolVar(&flags.diarize, "diarize", false, "Enable speaker identification")
	cmd.Flags().IntVar(&flags.speakers, "speakers", 0, "Max speakers (1-32, requires --diarize)")
	cmd.Flags().StringVar(&flags.timestamps, "timestamps", "word", "Timestamp granularity: none, word, character")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file (.txt, .srt, .json)")

	return cmd
}

func runSTT(cmd *cobra.Command, args []string, flags *sttFlags) error {
	// Get audio file from args, --file flag, or stdin
	audioFile, cleanup, err := getSTTAudioInput(args, flags.file, cmd.InOrStdin())
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return common.WriteError(cmd, "file_not_found", err.Error())
		}
		return common.WriteError(cmd, "missing_input", err.Error())
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Validate audio format
	ext := strings.ToLower(filepath.Ext(audioFile))
	if !supportedSTTFormats[ext] {
		return common.WriteError(cmd, "invalid_audio", fmt.Sprintf("unsupported audio format '%s', supported: mp3, wav, m4a, flac, ogg, webm, mp4, mkv, mov, avi", ext))
	}

	// Check file size
	info, err := os.Stat(audioFile)
	if err != nil {
		return common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot access file: %s", err.Error()))
	}
	if info.Size() > maxSTTFileSize {
		return common.WriteError(cmd, "file_too_large", fmt.Sprintf("file size exceeds 3GB limit"))
	}

	// Validate model
	if flags.model != "scribe_v1" && flags.model != "scribe_v2" {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s', supported: scribe_v1, scribe_v2", flags.model))
	}

	// Validate timestamps
	if flags.timestamps != "none" && flags.timestamps != "word" && flags.timestamps != "character" {
		return common.WriteError(cmd, "invalid_parameter", fmt.Sprintf("invalid timestamps value '%s', supported: none, word, character", flags.timestamps))
	}

	// Validate speakers requires diarize
	if flags.speakers > 0 && !flags.diarize {
		return common.WriteError(cmd, "invalid_parameter", "--speakers requires --diarize flag")
	}

	// Validate speakers range
	if flags.speakers < 0 || flags.speakers > 32 {
		return common.WriteError(cmd, "invalid_parameter", "--speakers must be between 1 and 32")
	}

	// Check API key
	apiKey := config.GetAPIKey("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ELEVENLABS_API_KEY"))
	}

	// Build multipart form request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file
	file, err := os.Open(audioFile)
	if err != nil {
		return common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot open file: %s", err.Error()))
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(audioFile))
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create form file: %s", err.Error()))
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot copy file: %s", err.Error()))
	}

	// Add model_id
	writer.WriteField("model_id", flags.model)

	// Add language if specified
	if flags.language != "" {
		writer.WriteField("language_code", flags.language)
	}

	// Add diarization
	if flags.diarize {
		writer.WriteField("diarize", "true")
		if flags.speakers > 0 {
			writer.WriteField("num_speakers", fmt.Sprintf("%d", flags.speakers))
		}
	}

	// Add timestamps granularity
	if flags.timestamps != "none" {
		writer.WriteField("timestamps_granularity", flags.timestamps)
	}

	writer.Close()

	// Make API request
	url := fmt.Sprintf("%s/speech-to-text", baseURL)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return handleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	// Handle API errors
	if resp.StatusCode != http.StatusOK {
		return handleAPIErrorResponse(cmd, resp)
	}

	// Parse response
	var apiResp sttAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Build response
	result := sttResponse{
		Success:  true,
		Text:     apiResp.Text,
		Language: apiResp.LanguageCode,
		Duration: apiResp.AudioDurationS,
	}

	// Add words if available (filter out spacing)
	if len(apiResp.Words) > 0 && flags.timestamps != "none" {
		for _, w := range apiResp.Words {
			if w.Type == "word" {
				result.Words = append(result.Words, sttWord{
					Word:  w.Text,
					Start: w.Start,
					End:   w.End,
				})
			}
		}
	}

	// Add speaker utterances if diarization was enabled
	if len(apiResp.Utterances) > 0 {
		result.Speakers = make([]sttSpeaker, len(apiResp.Utterances))
		for i, u := range apiResp.Utterances {
			result.Speakers[i] = sttSpeaker{
				Speaker: u.SpeakerID,
				Text:    u.Text,
				Start:   u.StartS,
				End:     u.EndS,
			}
		}
	}

	// Write to output file if specified
	if flags.output != "" {
		absPath, err := filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}

		ext := strings.ToLower(filepath.Ext(flags.output))
		var content string

		switch ext {
		case ".srt":
			content = generateSTTSRT(apiResp)
		case ".json":
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")
			content = string(jsonBytes)
		default: // .txt or other
			content = apiResp.Text
		}

		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
		}

		result.File = absPath
	}

	return common.WriteSuccess(cmd, result)
}

func getSTTAudioInput(args []string, filePath string, stdin io.Reader) (file string, cleanup func(), err error) {
	// From positional argument
	if len(args) > 0 && args[0] != "" {
		file := strings.TrimSpace(args[0])
		if file != "" {
			if _, err := os.Stat(file); err != nil {
				return "", nil, fmt.Errorf("audio file '%s' does not exist", file)
			}
			return file, nil, nil
		}
	}

	// From --file flag
	if filePath != "" {
		if _, err := os.Stat(filePath); err != nil {
			return "", nil, fmt.Errorf("audio file '%s' does not exist", filePath)
		}
		return filePath, nil, nil
	}

	// From stdin (only if not a terminal)
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", nil, errors.New("no audio file provided, use positional argument, --file flag, or pipe from stdin")
			}
		}

		// Read stdin into temp file
		tmpFile, err := os.CreateTemp("", "stt_stdin_*.mp3")
		if err != nil {
			return "", nil, fmt.Errorf("cannot create temp file: %w", err)
		}

		if _, err := io.Copy(tmpFile, stdin); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", nil, fmt.Errorf("cannot read stdin: %w", err)
		}

		tmpFile.Close()

		cleanup := func() {
			os.Remove(tmpFile.Name())
		}

		return tmpFile.Name(), cleanup, nil
	}

	return "", nil, errors.New("no audio file provided, use positional argument, --file flag, or pipe from stdin")
}

func generateSTTSRT(resp sttAPIResponse) string {
	// Filter to only word tokens
	var wordTokens []sttAPIWord
	for _, w := range resp.Words {
		if w.Type == "word" {
			wordTokens = append(wordTokens, w)
		}
	}

	if len(wordTokens) == 0 {
		return ""
	}

	var sb strings.Builder
	// Group words into segments (roughly 10 words per segment)
	const wordsPerSegment = 10
	segmentNum := 1

	for i := 0; i < len(wordTokens); i += wordsPerSegment {
		end := i + wordsPerSegment
		if end > len(wordTokens) {
			end = len(wordTokens)
		}

		segment := wordTokens[i:end]
		if len(segment) == 0 {
			continue
		}

		startTime := segment[0].Start
		endTime := segment[len(segment)-1].End

		var words []string
		for _, w := range segment {
			words = append(words, w.Text)
		}

		sb.WriteString(fmt.Sprintf("%d\n", segmentNum))
		sb.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTime(startTime), formatSRTTime(endTime)))
		sb.WriteString(strings.Join(words, " "))
		sb.WriteString("\n\n")
		segmentNum++
	}

	return sb.String()
}

func formatSRTTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds-float64(int(seconds)))*1000 + 0.5) // Round to nearest
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
