package google

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

// Supported audio formats
var sttSupportedFormats = map[string]string{
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".flac": "audio/flac",
	".ogg":  "audio/ogg",
	".aac":  "audio/aac",
	".m4a":  "audio/mp4",
	".webm": "audio/webm",
}

// Max file size (20 MB for inline, larger files use Files API)
const sttMaxInlineSize = 20 * 1024 * 1024

// STT response types
type sttResponse struct {
	Success  bool         `json:"success"`
	Text     string       `json:"text,omitempty"`
	Language string       `json:"language,omitempty"`
	Model    string       `json:"model,omitempty"`
	Segments []sttSegment `json:"segments,omitempty"`
	File     string       `json:"file,omitempty"`
}

type sttSegment struct {
	Speaker string `json:"speaker,omitempty"`
	Start   string `json:"start,omitempty"`
	End     string `json:"end,omitempty"`
	Text    string `json:"text"`
}

// Internal response from Gemini
type geminiSTTResponse struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
	Segments []struct {
		Speaker string `json:"speaker,omitempty"`
		Start   string `json:"start,omitempty"`
		End     string `json:"end,omitempty"`
		Text    string `json:"text"`
	} `json:"segments,omitempty"`
}

// STT flags
type sttFlags struct {
	file       string
	output     string
	language   string
	timestamps bool
	speakers   bool
	model      string
}

// Command
var sttCmd = newSTTCmd()

func newSTTCmd() *cobra.Command {
	flags := &sttFlags{}

	cmd := &cobra.Command{
		Use:           "stt [audio-file]",
		Short:         "Speech to Text using Google Gemini multimodal capabilities",
		Long:          "Transcribe audio files to text using Google Gemini multimodal capabilities.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSTT(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "Input audio file (alternative to positional arg)")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (prints to stdout if not set)")
	cmd.Flags().StringVarP(&flags.language, "language", "l", "", "Language hint (ISO 639-1 code, e.g., en, zh, ja)")
	cmd.Flags().BoolVarP(&flags.timestamps, "timestamps", "t", false, "Include timestamps in output")
	cmd.Flags().BoolVarP(&flags.speakers, "speakers", "s", false, "Enable speaker diarization")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "flash", "Model: flash")

	return cmd
}

func runSTT(cmd *cobra.Command, args []string, flags *sttFlags) error {
	// Get audio file path
	audioFile := ""
	if len(args) > 0 {
		audioFile = strings.TrimSpace(args[0])
	}
	if audioFile == "" && flags.file != "" {
		audioFile = flags.file
	}
	if audioFile == "" {
		return common.WriteError(cmd, "missing_input", "no audio file provided, use positional argument or --file flag")
	}

	// Check file exists
	info, err := os.Stat(audioFile)
	if err != nil {
		if os.IsNotExist(err) {
			return common.WriteError(cmd, "file_not_found", fmt.Sprintf("audio file '%s' does not exist", audioFile))
		}
		return common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot access file: %s", err.Error()))
	}

	// Validate audio format
	ext := strings.ToLower(filepath.Ext(audioFile))
	mimeType, ok := sttSupportedFormats[ext]
	if !ok {
		validFormats := make([]string, 0, len(sttSupportedFormats))
		for k := range sttSupportedFormats {
			validFormats = append(validFormats, k)
		}
		return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported audio format '%s', supported: %s", ext, strings.Join(validFormats, ", ")))
	}

	// Check file size (warn if too large, but still proceed with Files API)
	if info.Size() > 100*1024*1024 { // 100 MB limit
		return common.WriteError(cmd, "file_too_large", fmt.Sprintf("audio file size %d bytes exceeds 100 MB limit", info.Size()))
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

	// Build prompt for transcription
	prompt := buildTranscriptionPrompt(flags.language, flags.timestamps, flags.speakers)

	// Prepare audio content
	var parts []*genai.Part
	parts = append(parts, genai.NewPartFromText(prompt))

	// For smaller files, use inline data; for larger files, use Files API
	if info.Size() <= sttMaxInlineSize {
		// Read audio file and embed inline
		audioData, err := os.ReadFile(audioFile)
		if err != nil {
			return common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot read audio file: %s", err.Error()))
		}
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: mimeType,
				Data:     audioData,
			},
		})
	} else {
		// Upload file using Files API
		uploadedFile, err := client.Files.UploadFromPath(ctx, audioFile, &genai.UploadFileConfig{
			MIMEType: mimeType,
		})
		if err != nil {
			return common.WriteError(cmd, "upload_error", fmt.Sprintf("failed to upload audio file: %s", err.Error()))
		}
		// Schedule deletion after we're done
		defer func() {
			_, _ = client.Files.Delete(ctx, uploadedFile.Name, nil)
		}()

		parts = append(parts, genai.NewPartFromFile(*uploadedFile))
	}

	// Model ID
	modelID := "gemini-2.5-flash"

	// Build config
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	// Build content
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	// Call API
	result, err := client.Models.GenerateContent(ctx, modelID, contents, config)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Extract text from response
	responseText := ""
	if result.Candidates != nil && len(result.Candidates) > 0 {
		for _, part := range result.Candidates[0].Content.Parts {
			if part.Text != "" {
				responseText = part.Text
				break
			}
		}
	}

	if responseText == "" {
		return common.WriteError(cmd, "no_transcription", "no transcription generated in response")
	}

	// Parse the JSON response
	var geminiResp geminiSTTResponse
	if err := json.Unmarshal([]byte(responseText), &geminiResp); err != nil {
		// If parsing fails, treat the whole response as plain text
		geminiResp = geminiSTTResponse{Text: responseText}
	}

	// Build response
	resp := sttResponse{
		Success:  true,
		Text:     geminiResp.Text,
		Language: geminiResp.Language,
		Model:    modelID,
	}

	// Add segments if available
	if len(geminiResp.Segments) > 0 {
		resp.Segments = make([]sttSegment, len(geminiResp.Segments))
		for i, seg := range geminiResp.Segments {
			resp.Segments[i] = sttSegment{
				Speaker: seg.Speaker,
				Start:   seg.Start,
				End:     seg.End,
				Text:    seg.Text,
			}
		}
	}

	// Write to output file if specified
	if flags.output != "" {
		absPath, err := filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}

		output, _ := json.MarshalIndent(resp, "", "  ")
		if err := os.WriteFile(absPath, output, 0644); err != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
		}
		resp.File = absPath
	}

	return common.WriteSuccess(cmd, resp)
}

// buildTranscriptionPrompt creates the prompt for transcription
func buildTranscriptionPrompt(language string, timestamps, speakers bool) string {
	var sb strings.Builder

	sb.WriteString("Transcribe the following audio. ")

	if language != "" {
		sb.WriteString(fmt.Sprintf("The audio is in %s language. ", language))
	}

	sb.WriteString("Return the result as JSON with the following structure:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"text\": \"<full transcription>\",\n")
	sb.WriteString("  \"language\": \"<detected language code>\",\n")

	if timestamps || speakers {
		sb.WriteString("  \"segments\": [\n")
		sb.WriteString("    {\n")
		if speakers {
			sb.WriteString("      \"speaker\": \"<speaker identifier>\",\n")
		}
		if timestamps {
			sb.WriteString("      \"start\": \"<start time in MM:SS format>\",\n")
			sb.WriteString("      \"end\": \"<end time in MM:SS format>\",\n")
		}
		sb.WriteString("      \"text\": \"<segment text>\"\n")
		sb.WriteString("    }\n")
		sb.WriteString("  ]\n")
	}

	sb.WriteString("}\n")

	if timestamps {
		sb.WriteString("\nInclude timestamps for each segment.")
	}
	if speakers {
		sb.WriteString("\nIdentify different speakers and label them (Speaker 1, Speaker 2, etc.).")
	}

	return sb.String()
}
