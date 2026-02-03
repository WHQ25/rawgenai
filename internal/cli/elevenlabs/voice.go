package elevenlabs

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

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

// Voice Design command - design a new voice from description
type voiceDesignFlags struct {
	output         string
	description    string
	text           string
	autoText       bool
	model          string
	format         string
	loudness       float64
	guidanceScale  float64
	seed           int
	streamPreviews bool
	enhance        bool
	referenceAudio string
	promptStrength float64
	speak          bool
}

type voiceDesignResponse struct {
	Success  bool                   `json:"success"`
	File     string                 `json:"file,omitempty"`
	Previews []voicePreviewResponse `json:"previews"`
	Text     string                 `json:"text,omitempty"`
}

type voicePreviewResponse struct {
	GeneratedVoiceID string  `json:"generated_voice_id"`
	DurationSecs     float64 `json:"duration_secs"`
	Language         string  `json:"language,omitempty"`
}

type voiceDesignRequestBody struct {
	VoiceDescription    string   `json:"voice_description"`
	ModelID             string   `json:"model_id"`
	Text                string   `json:"text,omitempty"`
	AutoGenerateText    bool     `json:"auto_generate_text,omitempty"`
	Loudness            float64  `json:"loudness,omitempty"`
	GuidanceScale       float64  `json:"guidance_scale,omitempty"`
	Seed                *int     `json:"seed,omitempty"`
	StreamPreviews      bool     `json:"stream_previews,omitempty"`
	ShouldEnhance       bool     `json:"should_enhance,omitempty"`
	ReferenceAudioBase64 string  `json:"reference_audio_base64,omitempty"`
	PromptStrength      *float64 `json:"prompt_strength,omitempty"`
}

type voiceDesignAPIResponse struct {
	Previews []struct {
		AudioBase64      string  `json:"audio_base_64"`
		GeneratedVoiceID string  `json:"generated_voice_id"`
		MediaType        string  `json:"media_type"`
		DurationSecs     float64 `json:"duration_secs"`
		Language         string  `json:"language"`
	} `json:"previews"`
	Text string `json:"text"`
}

// Voice Create command - create a voice from preview
type voiceCreateFlags struct {
	name        string
	description string
	voiceID     string
	labels      string
}

type voiceCreateResponse struct {
	Success     bool              `json:"success"`
	VoiceID     string            `json:"voice_id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type voiceCreateRequestBody struct {
	VoiceName        string            `json:"voice_name"`
	VoiceDescription string            `json:"voice_description"`
	GeneratedVoiceID string            `json:"generated_voice_id"`
	Labels           map[string]string `json:"labels,omitempty"`
}

// Voice Preview command - stream a voice preview
type voicePreviewFlags struct {
	output string
	speak  bool
}

var voiceCmd = newVoiceCmd()

func newVoiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voice",
		Short: "Voice design and management commands",
		Long:  "Design new voices from text descriptions, create voices from previews, and stream voice previews.",
	}

	cmd.AddCommand(newVoiceDesignCmd())
	cmd.AddCommand(newVoiceCreateCmd())
	cmd.AddCommand(newVoicePreviewCmd())
	cmd.AddCommand(newVoiceListCmd())

	return cmd
}

func newVoiceDesignCmd() *cobra.Command {
	flags := &voiceDesignFlags{}

	cmd := &cobra.Command{
		Use:   "design [description] [flags]",
		Short: "Design a new voice from a text description",
		Example: `  rawgenai elevenlabs voice design "A warm, friendly female voice with a British accent" -o preview.mp3
  rawgenai elevenlabs voice design "Deep male narrator" --text "Hello, welcome to our story" -o narrator.mp3
  rawgenai elevenlabs voice design "Energetic young voice" --auto-text -o preview.mp3
  rawgenai elevenlabs voice design "Similar to this" --reference-audio sample.mp3 -o preview.mp3`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVoiceDesign(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path for first preview (.mp3)")
	cmd.Flags().StringVar(&flags.description, "description", "", "Voice description (alternative to positional arg)")
	cmd.Flags().StringVar(&flags.text, "text", "", "Text to speak in preview (100-1000 chars)")
	cmd.Flags().BoolVar(&flags.autoText, "auto-text", false, "Auto-generate preview text")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "eleven_multilingual_ttv_v2", "Model: eleven_multilingual_ttv_v2, eleven_ttv_v3")
	cmd.Flags().StringVarP(&flags.format, "format", "f", "mp3_44100_128", "Output format")
	cmd.Flags().Float64Var(&flags.loudness, "loudness", 0.5, "Volume level (-1 to 1)")
	cmd.Flags().Float64Var(&flags.guidanceScale, "guidance-scale", 5.0, "How closely AI follows prompt (1-20)")
	cmd.Flags().IntVar(&flags.seed, "seed", 0, "Random seed for reproducibility")
	cmd.Flags().BoolVar(&flags.streamPreviews, "stream-previews", false, "Return only IDs for streaming")
	cmd.Flags().BoolVar(&flags.enhance, "enhance", false, "AI-enhance the voice description")
	cmd.Flags().StringVar(&flags.referenceAudio, "reference-audio", "", "Reference audio file (eleven_ttv_v3 only)")
	cmd.Flags().Float64Var(&flags.promptStrength, "prompt-strength", 0.5, "Prompt vs reference balance (0-1, eleven_ttv_v3 only)")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play first preview after generation")

	return cmd
}

func runVoiceDesign(cmd *cobra.Command, args []string, flags *voiceDesignFlags) error {
	// Get description from args or flag
	description := flags.description
	if len(args) > 0 {
		description = strings.Join(args, " ")
	}
	if description == "" {
		return common.WriteError(cmd, "missing_description", "voice description is required")
	}

	// Validate text length if provided
	if flags.text != "" && (len(flags.text) < 100 || len(flags.text) > 1000) {
		return common.WriteError(cmd, "invalid_text_length", "text must be between 100 and 1000 characters")
	}

	// Validate loudness
	if flags.loudness < -1.0 || flags.loudness > 1.0 {
		return common.WriteError(cmd, "invalid_loudness", "loudness must be between -1.0 and 1.0")
	}

	// Validate guidance scale
	if flags.guidanceScale < 1.0 || flags.guidanceScale > 20.0 {
		return common.WriteError(cmd, "invalid_guidance_scale", "guidance scale must be between 1.0 and 20.0")
	}

	// Validate format
	if !ttsOutputFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("unsupported format '%s'", flags.format))
	}

	// Validate reference audio is only used with v3 model
	if flags.referenceAudio != "" && flags.model != "eleven_ttv_v3" {
		return common.WriteError(cmd, "incompatible_reference_audio", "reference audio requires eleven_ttv_v3 model")
	}

	// Check API key
	apiKey := config.GetAPIKey("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ELEVENLABS_API_KEY"))
	}

	// Build request body
	reqBody := voiceDesignRequestBody{
		VoiceDescription: description,
		ModelID:          flags.model,
		Loudness:         flags.loudness,
		GuidanceScale:    flags.guidanceScale,
		StreamPreviews:   flags.streamPreviews,
		ShouldEnhance:    flags.enhance,
	}

	if flags.text != "" {
		reqBody.Text = flags.text
	}
	if flags.autoText {
		reqBody.AutoGenerateText = true
	}
	if flags.seed > 0 {
		reqBody.Seed = &flags.seed
	}

	// Handle reference audio
	if flags.referenceAudio != "" {
		audioData, err := os.ReadFile(flags.referenceAudio)
		if err != nil {
			return common.WriteError(cmd, "file_read_error", fmt.Sprintf("cannot read reference audio: %s", err.Error()))
		}
		reqBody.ReferenceAudioBase64 = base64.StdEncoding.EncodeToString(audioData)
		reqBody.PromptStrength = &flags.promptStrength
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	// Make API request
	apiURL := fmt.Sprintf("%s/text-to-voice/design?output_format=%s", baseURL, flags.format)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

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
	var apiResp voiceDesignAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "invalid_response", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Build result
	result := voiceDesignResponse{
		Success:  true,
		Text:     apiResp.Text,
		Previews: make([]voicePreviewResponse, len(apiResp.Previews)),
	}

	for i, p := range apiResp.Previews {
		result.Previews[i] = voicePreviewResponse{
			GeneratedVoiceID: p.GeneratedVoiceID,
			DurationSecs:     p.DurationSecs,
			Language:         p.Language,
		}
	}

	// Save first preview to file if output specified
	if flags.output != "" && len(apiResp.Previews) > 0 && apiResp.Previews[0].AudioBase64 != "" {
		audioData, err := base64.StdEncoding.DecodeString(apiResp.Previews[0].AudioBase64)
		if err != nil {
			return common.WriteError(cmd, "decode_error", fmt.Sprintf("cannot decode audio: %s", err.Error()))
		}

		absPath, err := filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}

		if err := os.WriteFile(absPath, audioData, 0644); err != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
		}
		result.File = absPath

		// Play if --speak
		if flags.speak {
			if err := common.PlayFile(absPath); err != nil {
				return common.WriteError(cmd, "playback_error", fmt.Sprintf("cannot play audio: %s", err.Error()))
			}
		}
	}

	return common.WriteSuccess(cmd, result)
}

func newVoiceCreateCmd() *cobra.Command {
	flags := &voiceCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create [flags]",
		Short: "Create a voice from a generated preview",
		Example: `  rawgenai elevenlabs voice create --name "My Voice" --description "A custom voice" --voice-id abc123
  rawgenai elevenlabs voice create -n "Narrator" -d "Deep narrator voice" --voice-id xyz789 --labels '{"accent":"british"}'`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVoiceCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.name, "name", "n", "", "Name for the new voice (required)")
	cmd.Flags().StringVarP(&flags.description, "description", "d", "", "Description for the new voice (required)")
	cmd.Flags().StringVar(&flags.voiceID, "voice-id", "", "Generated voice ID from design command (required)")
	cmd.Flags().StringVar(&flags.labels, "labels", "", "JSON object with voice labels")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("description")
	cmd.MarkFlagRequired("voice-id")

	return cmd
}

func runVoiceCreate(cmd *cobra.Command, args []string, flags *voiceCreateFlags) error {
	// Validate required fields
	if flags.name == "" {
		return common.WriteError(cmd, "missing_name", "voice name is required")
	}
	if flags.description == "" {
		return common.WriteError(cmd, "missing_description", "voice description is required")
	}
	if flags.voiceID == "" {
		return common.WriteError(cmd, "missing_voice_id", "generated voice ID is required")
	}

	// Parse labels if provided
	var labels map[string]string
	if flags.labels != "" {
		if err := json.Unmarshal([]byte(flags.labels), &labels); err != nil {
			return common.WriteError(cmd, "invalid_labels", fmt.Sprintf("invalid JSON labels: %s", err.Error()))
		}
	}

	// Check API key
	apiKey := config.GetAPIKey("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ELEVENLABS_API_KEY"))
	}

	// Build request body
	reqBody := voiceCreateRequestBody{
		VoiceName:        flags.name,
		VoiceDescription: flags.description,
		GeneratedVoiceID: flags.voiceID,
		Labels:           labels,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	// Make API request
	apiURL := fmt.Sprintf("%s/text-to-voice", baseURL)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

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
	var apiResp struct {
		VoiceID     string            `json:"voice_id"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Labels      map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "invalid_response", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	return common.WriteSuccess(cmd, voiceCreateResponse{
		Success:     true,
		VoiceID:     apiResp.VoiceID,
		Name:        apiResp.Name,
		Description: apiResp.Description,
		Labels:      apiResp.Labels,
	})
}

func newVoicePreviewCmd() *cobra.Command {
	flags := &voicePreviewFlags{}

	cmd := &cobra.Command{
		Use:   "preview [generated_voice_id] [flags]",
		Short: "Stream a voice preview by ID",
		Example: `  rawgenai elevenlabs voice preview abc123 -o preview.mp3
  rawgenai elevenlabs voice preview abc123 --speak`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVoicePreview(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp3)")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after download")

	return cmd
}

func runVoicePreview(cmd *cobra.Command, args []string, flags *voicePreviewFlags) error {
	// Get voice ID from args
	if len(args) == 0 {
		return common.WriteError(cmd, "missing_voice_id", "generated voice ID is required as first argument")
	}
	voiceID := args[0]

	// Validate output or speak
	if flags.output == "" && !flags.speak {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag or --speak")
	}

	// Determine output path
	var outputPath string
	var useTempFile bool

	if flags.output != "" {
		outputPath = flags.output
	} else {
		tmpFile, err := os.CreateTemp("", "voice-preview-*.mp3")
		if err != nil {
			return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create temp file: %s", err.Error()))
		}
		outputPath = tmpFile.Name()
		tmpFile.Close()
		useTempFile = true
	}

	// Check API key
	apiKey := config.GetAPIKey("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ELEVENLABS_API_KEY"))
	}

	// Make API request
	apiURL := fmt.Sprintf("%s/text-to-voice/%s/stream", baseURL, voiceID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("xi-api-key", apiKey)

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

	// Get absolute path
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

	// Play if --speak
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

	result := struct {
		Success bool   `json:"success"`
		File    string `json:"file,omitempty"`
		VoiceID string `json:"voice_id"`
	}{
		Success: true,
		File:    absPath,
		VoiceID: voiceID,
	}
	if useTempFile {
		result.File = ""
	}
	return common.WriteSuccess(cmd, result)
}
