package elevenlabs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type dialogueFlags struct {
	output            string
	inputFile         string
	model             string
	format            string
	language          string
	stability         float64
	textNormalization string
	seed              int
	speak             bool
}

type dialogueResponse struct {
	Success  bool   `json:"success"`
	File     string `json:"file,omitempty"`
	Model    string `json:"model,omitempty"`
	Segments int    `json:"segments,omitempty"`
}

type dialogueInput struct {
	Text    string `json:"text"`
	VoiceID string `json:"voice_id"`
}

type dialogueRequestBody struct {
	Inputs                 []dialogueInput   `json:"inputs"`
	ModelID                string            `json:"model_id"`
	LanguageCode           string            `json:"language_code,omitempty"`
	Settings               *dialogueSettings `json:"settings,omitempty"`
	ApplyTextNormalization string            `json:"apply_text_normalization,omitempty"`
	Seed                   *int              `json:"seed,omitempty"`
}

type dialogueSettings struct {
	Stability float64 `json:"stability"`
}

var dialogueCmd = newDialogueCmd()

func newDialogueCmd() *cobra.Command {
	flags := &dialogueFlags{}

	cmd := &cobra.Command{
		Use:   "dialogue [flags]",
		Short: "Generate multi-voice dialogue from text",
		Long: `Generate multi-voice dialogue by providing a JSON array of text and voice ID pairs.
The input JSON should be an array of objects with "text" and "voice_id" fields.
Maximum 10 unique voices per request.

The voice_id can be:
  - A preset name: rachel, josh, bella, adam, sam, charlie, arnold, elli, domi, antoni
  - An actual voice ID from "rawgenai elevenlabs voice list"`,
		Example: `  # Using preset voice names
  echo '[{"text":"Hello","voice_id":"rachel"},{"text":"Hi","voice_id":"josh"}]' | \
    rawgenai elevenlabs dialogue -o chat.mp3

  # Using actual voice IDs (from "voice list")
  echo '[{"text":"Hello","voice_id":"CwhRBWXzGAHq8TQ4Fs17"}]' | \
    rawgenai elevenlabs dialogue -o chat.mp3

  # Using input file
  rawgenai elevenlabs dialogue -i dialogue.json -o conversation.mp3`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDialogue(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp3)")
	cmd.Flags().StringVarP(&flags.inputFile, "input", "i", "", "Input JSON file with dialogue data")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "eleven_v3", "Model: eleven_v3, eleven_multilingual_v2")
	cmd.Flags().StringVarP(&flags.format, "format", "f", "mp3_44100_128", "Output format")
	cmd.Flags().StringVarP(&flags.language, "language", "l", "", "Language code (ISO 639-1)")
	cmd.Flags().Float64Var(&flags.stability, "stability", 0.5, "Voice stability (0.0-1.0)")
	cmd.Flags().StringVar(&flags.textNormalization, "text-normalization", "auto", "Text normalization: auto, on, off")
	cmd.Flags().IntVar(&flags.seed, "seed", 0, "Random seed for deterministic generation")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	return cmd
}

func runDialogue(cmd *cobra.Command, args []string, flags *dialogueFlags) error {
	// Read dialogue input from file or stdin
	var inputData []byte
	var err error

	if flags.inputFile != "" {
		inputData, err = os.ReadFile(flags.inputFile)
		if err != nil {
			return common.WriteError(cmd, "file_read_error", fmt.Sprintf("cannot read input file: %s", err.Error()))
		}
	} else {
		// Read from stdin
		inputData, err = io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return common.WriteError(cmd, "stdin_read_error", fmt.Sprintf("cannot read stdin: %s", err.Error()))
		}
	}

	if len(inputData) == 0 {
		return common.WriteError(cmd, "missing_input", "no dialogue input provided, use -i flag or pipe JSON to stdin")
	}

	// Parse dialogue inputs
	var rawInputs []struct {
		Text    string `json:"text"`
		VoiceID string `json:"voice_id"`
	}
	if err := json.Unmarshal(inputData, &rawInputs); err != nil {
		return common.WriteError(cmd, "invalid_input", fmt.Sprintf("invalid JSON input: %s", err.Error()))
	}

	if len(rawInputs) == 0 {
		return common.WriteError(cmd, "empty_input", "dialogue input array is empty")
	}

	// Resolve voice IDs and build inputs
	inputs := make([]dialogueInput, len(rawInputs))
	uniqueVoices := make(map[string]bool)
	for i, raw := range rawInputs {
		if raw.Text == "" {
			return common.WriteError(cmd, "empty_text", fmt.Sprintf("text is empty at index %d", i))
		}
		voiceID := resolveVoiceID(raw.VoiceID)
		inputs[i] = dialogueInput{
			Text:    raw.Text,
			VoiceID: voiceID,
		}
		uniqueVoices[voiceID] = true
	}

	if len(uniqueVoices) > 10 {
		return common.WriteError(cmd, "too_many_voices", "maximum 10 unique voices allowed per request")
	}

	// Validate output or speak
	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	// Determine output path
	var outputPath string
	var useTempFile bool
	outputFormat := flags.format

	if flags.output != "" {
		outputPath = flags.output
	} else {
		outputFormat = "mp3_44100_128"
		tmpFile, err := os.CreateTemp("", "dialogue-*.mp3")
		if err != nil {
			return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create temp file: %s", err.Error()))
		}
		outputPath = tmpFile.Name()
		tmpFile.Close()
		useTempFile = true
	}

	// Validate format
	if !ttsOutputFormats[outputFormat] {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("unsupported format '%s'", outputFormat))
	}

	// Validate text normalization
	validTextNorm := map[string]bool{"auto": true, "on": true, "off": true}
	if !validTextNorm[flags.textNormalization] {
		return common.WriteError(cmd, "invalid_text_normalization", "text normalization must be auto, on, or off")
	}

	// Validate stability
	if flags.stability < 0.0 || flags.stability > 1.0 {
		return common.WriteError(cmd, "invalid_stability", "stability must be between 0.0 and 1.0")
	}

	// Check API key
	apiKey := config.GetAPIKey("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ELEVENLABS_API_KEY"))
	}

	// Build request body
	reqBody := dialogueRequestBody{
		Inputs:  inputs,
		ModelID: flags.model,
	}

	if flags.language != "" {
		reqBody.LanguageCode = flags.language
	}

	if flags.stability != 0.5 {
		reqBody.Settings = &dialogueSettings{
			Stability: flags.stability,
		}
	}

	if flags.textNormalization != "auto" {
		reqBody.ApplyTextNormalization = flags.textNormalization
	}

	if flags.seed > 0 {
		reqBody.Seed = &flags.seed
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	// Make API request
	apiURL := fmt.Sprintf("%s/text-to-dialogue?output_format=%s", baseURL, outputFormat)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
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

	// Write response to file
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
		outFile.Close()
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
	result := dialogueResponse{
		Success:  true,
		File:     absPath,
		Model:    flags.model,
		Segments: len(inputs),
	}
	if useTempFile {
		result.File = ""
	}
	return common.WriteSuccess(cmd, result)
}
