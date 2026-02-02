package google

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

// TTS model mapping
var ttsModelIDs = map[string]string{
	"flash": "gemini-2.5-flash-preview-tts",
	"pro":   "gemini-2.5-pro-preview-tts",
}

// Valid voice names
var validVoices = map[string]bool{
	"Zephyr": true, "Puck": true, "Charon": true, "Kore": true, "Fenrir": true,
	"Leda": true, "Orus": true, "Aoede": true, "Callirrhoe": true, "Autonoe": true,
	"Enceladus": true, "Iapetus": true, "Umbriel": true, "Algieba": true, "Despina": true,
	"Erinome": true, "Algenib": true, "Rasalgethi": true, "Laomedeia": true, "Achernar": true,
	"Alnilam": true, "Schedar": true, "Gacrux": true, "Pulcherrima": true, "Achird": true,
	"Zubenelgenubi": true, "Vindemiatrix": true, "Sadachbia": true, "Sadaltager": true, "Sulafat": true,
}

// TTS response types
type ttsResponse struct {
	Success  bool              `json:"success"`
	File     string            `json:"file,omitempty"`
	Model    string            `json:"model,omitempty"`
	Voice    string            `json:"voice,omitempty"`
	Speakers map[string]string `json:"speakers,omitempty"`
}

// TTS flags
type ttsFlags struct {
	output     string
	promptFile string
	voice      string
	speakers   string
	model      string
	speak      bool
}

// Command
var ttsCmd = newTTSCmd()

func newTTSCmd() *cobra.Command {
	flags := &ttsFlags{}

	cmd := &cobra.Command{
		Use:   "tts [prompt]",
		Short: "Text to Speech using Google Gemini TTS models",
		Long: `Convert text to speech using Google Gemini TTS models (flash and pro).

Unlike other TTS providers, Gemini TTS accepts a prompt that can include
style instructions to control voice characteristics like tone, pace, and emotion.

Examples:
  rawgenai google tts "Hello world" -o hello.wav
  rawgenai google tts "Say cheerfully: Hello everyone!" -o cheerful.wav
  rawgenai google tts "Whisper: This is a secret" -o whisper.wav`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTTS(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.wav)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")
	cmd.Flags().StringVarP(&flags.voice, "voice", "v", "Kore", "Voice name (single speaker)")
	cmd.Flags().StringVar(&flags.speakers, "speakers", "", "Multi-speaker config: \"Name1=Voice1,Name2=Voice2\"")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "flash", "Model: flash, pro")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	return cmd
}

func runTTS(cmd *cobra.Command, args []string, flags *ttsFlags) error {
	// Get text from args, file, or stdin
	text, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_text", err.Error())
	}

	// Validate output or speak
	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	// Determine output path
	var outputPath string
	var useTempFile bool

	if flags.output != "" {
		outputPath = flags.output
		// Validate format (only WAV supported)
		ext := strings.ToLower(filepath.Ext(outputPath))
		if ext != ".wav" {
			return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s', only .wav is supported", ext))
		}
	} else {
		// --speak only: use temp file
		tmpFile, err := os.CreateTemp("", "tts-*.wav")
		if err != nil {
			return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create temp file: %s", err.Error()))
		}
		outputPath = tmpFile.Name()
		tmpFile.Close()
		useTempFile = true
	}

	// Validate model
	modelID, ok := ttsModelIDs[flags.model]
	if !ok {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s', use 'flash' or 'pro'", flags.model))
	}

	// Parse speakers config if provided
	var speakerMap map[string]string
	if flags.speakers != "" {
		if flags.voice != "Kore" {
			return common.WriteError(cmd, "conflicting_flags", "cannot use --voice and --speakers together")
		}
		speakerMap, err = parseSpeakers(flags.speakers)
		if err != nil {
			return common.WriteError(cmd, "invalid_speakers", err.Error())
		}
		if len(speakerMap) > 2 {
			return common.WriteError(cmd, "too_many_speakers", "maximum 2 speakers allowed")
		}
	} else {
		// Validate single voice
		if !validVoices[flags.voice] {
			return common.WriteError(cmd, "invalid_voice", fmt.Sprintf("voice '%s' is not a valid prebuilt voice", flags.voice))
		}
	}

	// Check API key
	apiKey := config.GetAPIKey("GEMINI_API_KEY", "GOOGLE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("GEMINI_API_KEY", "GOOGLE_API_KEY"))
	}

	// Create client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return common.WriteError(cmd, "client_error", fmt.Sprintf("failed to create client: %s", err.Error()))
	}

	// Build config
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO"},
		SpeechConfig:       &genai.SpeechConfig{},
	}

	// Configure voice
	if speakerMap != nil {
		// Multi-speaker mode
		var speakerConfigs []*genai.SpeakerVoiceConfig
		for speaker, voice := range speakerMap {
			speakerConfigs = append(speakerConfigs, &genai.SpeakerVoiceConfig{
				Speaker: speaker,
				VoiceConfig: &genai.VoiceConfig{
					PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
						VoiceName: voice,
					},
				},
			})
		}
		config.SpeechConfig.MultiSpeakerVoiceConfig = &genai.MultiSpeakerVoiceConfig{
			SpeakerVoiceConfigs: speakerConfigs,
		}
	} else {
		// Single speaker mode
		config.SpeechConfig.VoiceConfig = &genai.VoiceConfig{
			PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
				VoiceName: flags.voice,
			},
		}
	}

	// Call API
	result, err := client.Models.GenerateContent(ctx, modelID, genai.Text(text), config)
	if err != nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return handleAPIError(cmd, err)
	}

	// Extract audio from response
	var audioBytes []byte
	if result.Candidates != nil && len(result.Candidates) > 0 && result.Candidates[0].Content != nil {
		for _, part := range result.Candidates[0].Content.Parts {
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "audio/") {
				audioBytes = part.InlineData.Data
				break
			}
		}
	}

	if audioBytes == nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "no_audio", "no audio generated in response")
	}

	// Convert PCM to WAV
	wavBytes := pcmToWAV(audioBytes, 24000, 16, 1)

	// Save audio
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	if err := os.WriteFile(absPath, wavBytes, 0644); err != nil {
		if useTempFile {
			os.Remove(outputPath)
		}
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	// Play audio if --speak is set
	if flags.speak {
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

	// Build response
	resp := ttsResponse{
		Success: true,
		File:    absPath,
		Model:   modelID,
	}
	if useTempFile {
		resp.File = "" // Don't report temp file path
	}
	if speakerMap != nil {
		resp.Speakers = speakerMap
	} else {
		resp.Voice = flags.voice
	}

	return common.WriteSuccess(cmd, resp)
}

// parseSpeakers parses the speaker config string
// Format: "Name1=Voice1,Name2=Voice2"
func parseSpeakers(config string) (map[string]string, error) {
	result := make(map[string]string)
	pairs := strings.Split(config, ",")

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid speaker format '%s', expected 'Name=Voice'", pair)
		}
		speaker := strings.TrimSpace(parts[0])
		voice := strings.TrimSpace(parts[1])

		if speaker == "" {
			return nil, errors.New("speaker name cannot be empty")
		}
		if !validVoices[voice] {
			return nil, fmt.Errorf("voice '%s' is not a valid prebuilt voice", voice)
		}
		result[speaker] = voice
	}

	if len(result) == 0 {
		return nil, errors.New("no valid speakers specified")
	}

	return result, nil
}

// pcmToWAV converts raw PCM audio to WAV format
// PCM format: 16-bit signed little-endian
func pcmToWAV(pcm []byte, sampleRate, bitsPerSample, channels int) []byte {
	// WAV header is 44 bytes
	dataSize := len(pcm)
	fileSize := dataSize + 36 // Total file size minus 8 bytes for RIFF header

	header := make([]byte, 44)

	// RIFF header
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], uint32(fileSize))
	copy(header[8:12], "WAVE")

	// fmt subchunk
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16)                                              // Subchunk1Size (16 for PCM)
	binary.LittleEndian.PutUint16(header[20:22], 1)                                               // AudioFormat (1 for PCM)
	binary.LittleEndian.PutUint16(header[22:24], uint16(channels))                                // NumChannels
	binary.LittleEndian.PutUint32(header[24:28], uint32(sampleRate))                              // SampleRate
	binary.LittleEndian.PutUint32(header[28:32], uint32(sampleRate*channels*bitsPerSample/8))     // ByteRate
	binary.LittleEndian.PutUint16(header[32:34], uint16(channels*bitsPerSample/8))                // BlockAlign
	binary.LittleEndian.PutUint16(header[34:36], uint16(bitsPerSample))                           // BitsPerSample

	// data subchunk
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], uint32(dataSize))

	// Combine header and data
	wav := make([]byte, 44+dataSize)
	copy(wav[0:44], header)
	copy(wav[44:], pcm)

	return wav
}

// Helper to get text from various sources (reusing from image.go pattern)
func getTTSText(args []string, filePath string, stdin io.Reader) (string, error) {
	// Priority 1: Positional argument
	if len(args) > 0 {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text != "" {
			return text, nil
		}
	}

	// Priority 2: File
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("file not found: %s", filePath)
			}
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", errors.New("file is empty")
		}
		return text, nil
	}

	// Priority 3: Stdin (only if not a terminal)
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", errors.New("no text provided")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", errors.New("stdin is empty")
		}
		return text, nil
	}

	return "", errors.New("no text provided")
}
