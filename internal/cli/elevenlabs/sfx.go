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
	"github.com/spf13/cobra"
)

var sfxOutputFormats = map[string]bool{
	"mp3_22050_32":  true,
	"mp3_44100_128": true,
	"mp3_44100_192": true,
	"pcm_44100":     true,
	"pcm_48000":     true,
}

type sfxFlags struct {
	output     string
	promptFile string
	duration   float64
	loop       bool
	influence  float64
	format     string
}

type sfxResponse struct {
	Success  bool    `json:"success"`
	File     string  `json:"file,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Loop     bool    `json:"loop,omitempty"`
}

type sfxRequestBody struct {
	Text             string  `json:"text"`
	DurationSeconds  float64 `json:"duration_seconds,omitempty"`
	PromptInfluence  float64 `json:"prompt_influence,omitempty"`
}

var sfxCmd = newSFXCmd()

func newSFXCmd() *cobra.Command {
	flags := &sfxFlags{}

	cmd := &cobra.Command{
		Use:           "sfx [prompt]",
		Short:         "Generate sound effects using ElevenLabs",
		Long:          "Generate sound effects from text prompts using ElevenLabs sound generation API.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSFX(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp3)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")
	cmd.Flags().Float64VarP(&flags.duration, "duration", "d", 0, "Duration in seconds (0.5-30, 0 for auto)")
	cmd.Flags().BoolVar(&flags.loop, "loop", false, "Create seamless loop")
	cmd.Flags().Float64Var(&flags.influence, "influence", 0.3, "Prompt influence (0.0-1.0)")
	cmd.Flags().StringVarP(&flags.format, "format", "f", "mp3_44100_128", "Output format")

	return cmd
}

func runSFX(cmd *cobra.Command, args []string, flags *sfxFlags) error {
	// Get prompt from args, file, or stdin
	prompt, err := getText(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate format
	if !sfxOutputFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("unsupported format '%s', supported: mp3_22050_32, mp3_44100_128, mp3_44100_192, pcm_44100, pcm_48000", flags.format))
	}

	// Validate duration (0 means auto)
	if flags.duration != 0 && (flags.duration < 0.5 || flags.duration > 30) {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 0.5 and 30 seconds (or 0 for auto)")
	}

	// Validate influence
	if flags.influence < 0.0 || flags.influence > 1.0 {
		return common.WriteError(cmd, "invalid_influence", "influence must be between 0.0 and 1.0")
	}

	// Check API key
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "ELEVENLABS_API_KEY environment variable is not set")
	}

	// Build request body
	reqBody := sfxRequestBody{
		Text:            prompt,
		PromptInfluence: flags.influence,
	}

	if flags.duration > 0 {
		reqBody.DurationSeconds = flags.duration
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	// Make API request
	url := fmt.Sprintf("%s/sound-generation?output_format=%s", baseURL, flags.format)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
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
	result := sfxResponse{
		Success:  true,
		File:     absPath,
		Duration: flags.duration,
		Loop:     flags.loop,
	}
	return common.WriteSuccess(cmd, result)
}
