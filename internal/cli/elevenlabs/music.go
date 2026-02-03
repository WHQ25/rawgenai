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

type musicFlags struct {
	output           string
	promptFile       string
	duration         int
	instrumental     bool
	format           string
	compositionPlan  string
	respectDurations bool
	speak            bool
}

type musicResponse struct {
	Success      bool   `json:"success"`
	File         string `json:"file,omitempty"`
	DurationMs   int    `json:"duration_ms,omitempty"`
	Instrumental bool   `json:"instrumental,omitempty"`
}

type musicRequestBody struct {
	Prompt                   string       `json:"prompt,omitempty"`
	CompositionPlan          *musicPrompt `json:"composition_plan,omitempty"`
	MusicLengthMs            *int         `json:"music_length_ms,omitempty"`
	ModelID                  string       `json:"model_id"`
	ForceInstrumental        bool         `json:"force_instrumental,omitempty"`
	RespectSectionsDurations bool         `json:"respect_sections_durations,omitempty"`
}

type musicPrompt struct {
	PositiveGlobalStyles []string      `json:"positive_global_styles"`
	NegativeGlobalStyles []string      `json:"negative_global_styles"`
	Sections             []songSection `json:"sections"`
}

type songSection struct {
	SectionName          string   `json:"section_name"`
	PositiveLocalStyles  []string `json:"positive_local_styles"`
	NegativeLocalStyles  []string `json:"negative_local_styles"`
	DurationMs           int      `json:"duration_ms"`
	Lines                []string `json:"lines"`
}

var musicCmd = newMusicCmd()

func newMusicCmd() *cobra.Command {
	flags := &musicFlags{}

	cmd := &cobra.Command{
		Use:   "music [prompt] [flags]",
		Short: "Generate music from a text prompt",
		Example: `  rawgenai elevenlabs music "upbeat electronic dance track" -o track.mp3
  rawgenai elevenlabs music "calm piano melody" -o piano.mp3 -d 60000
  rawgenai elevenlabs music "rock anthem" -o rock.mp3 --instrumental
  rawgenai elevenlabs music --composition-plan plan.json -o song.mp3
  echo "jazz fusion" | rawgenai elevenlabs music -o jazz.mp3`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMusic(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.mp3)")
	cmd.Flags().StringVar(&flags.promptFile, "file", "", "Input prompt file")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 0, "Music length in milliseconds (3000-600000)")
	cmd.Flags().BoolVar(&flags.instrumental, "instrumental", false, "Force instrumental (no vocals)")
	cmd.Flags().StringVarP(&flags.format, "format", "f", "mp3_44100_128", "Output format")
	cmd.Flags().StringVar(&flags.compositionPlan, "composition-plan", "", "JSON file with detailed composition plan")
	cmd.Flags().BoolVar(&flags.respectDurations, "respect-durations", true, "Strictly respect section durations in composition plan")
	cmd.Flags().BoolVar(&flags.speak, "speak", false, "Play audio after generation")

	return cmd
}

func runMusic(cmd *cobra.Command, args []string, flags *musicFlags) error {
	// Check for composition plan first
	var compositionPlan *musicPrompt
	var prompt string
	var err error

	if flags.compositionPlan != "" {
		// Read composition plan from file
		planData, err := os.ReadFile(flags.compositionPlan)
		if err != nil {
			return common.WriteError(cmd, "file_read_error", fmt.Sprintf("cannot read composition plan: %s", err.Error()))
		}
		if err := json.Unmarshal(planData, &compositionPlan); err != nil {
			return common.WriteError(cmd, "invalid_composition_plan", fmt.Sprintf("invalid JSON in composition plan: %s", err.Error()))
		}
	} else {
		// Get prompt from args, file, or stdin
		prompt, err = getText(args, flags.promptFile, cmd.InOrStdin())
		if err != nil {
			return common.WriteError(cmd, "missing_prompt", err.Error())
		}
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
		// --speak only: use temp file
		outputFormat = "mp3_44100_128"
		tmpFile, err := os.CreateTemp("", "music-*.mp3")
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

	// Validate duration
	if flags.duration != 0 && (flags.duration < 3000 || flags.duration > 600000) {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 3000 and 600000 milliseconds")
	}

	// Check API key
	apiKey := config.GetAPIKey("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ELEVENLABS_API_KEY"))
	}

	// Build request body
	reqBody := musicRequestBody{
		ModelID: "music_v1",
	}

	if compositionPlan != nil {
		reqBody.CompositionPlan = compositionPlan
		reqBody.RespectSectionsDurations = flags.respectDurations
	} else {
		reqBody.Prompt = prompt
		if flags.duration > 0 {
			reqBody.MusicLengthMs = &flags.duration
		}
		reqBody.ForceInstrumental = flags.instrumental
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	// Make API request
	apiURL := fmt.Sprintf("%s/music?output_format=%s", baseURL, outputFormat)
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
	result := musicResponse{
		Success:      true,
		File:         absPath,
		DurationMs:   flags.duration,
		Instrumental: flags.instrumental,
	}
	if useTempFile {
		result.File = ""
	}
	return common.WriteSuccess(cmd, result)
}
