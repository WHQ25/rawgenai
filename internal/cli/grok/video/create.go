package video

import (
	"bytes"
	"encoding/base64"
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

const videoGenerationsPath = "/videos/generations"

// Valid aspect ratios for video generation
var validVideoAspects = map[string]bool{
	"16:9": true,
	"9:16": true,
}

// Valid resolutions
var validResolutions = map[string]bool{
	"720p": true,
	"480p": true,
}

// Valid image formats for input
var validImageFormats = map[string]string{
	".png":  "image/png",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
}

type createFlags struct {
	promptFile string
	image      string
	duration   int
	aspect     string
	resolution string
}

type createResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// API response type
type xaiVideoCreateResponse struct {
	RequestID string         `json:"request_id"`
	Status    string         `json:"status"`
	Error     *xaiVideoError `json:"error,omitempty"`
}

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:   "create [prompt]",
		Short: "Create a video generation job",
		Long: `Create a video generation job and return the request ID.

Use 'video status' to check progress and 'video download' to retrieve the result.

IMPORTANT: Save the request_id from the response. You need it to check status
and download the video.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read prompt from file")
	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Input image for image-to-video")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 5, "Duration in seconds (1-15)")
	cmd.Flags().StringVarP(&flags.aspect, "aspect", "a", "16:9", "Aspect ratio: 16:9, 9:16")
	cmd.Flags().StringVarP(&flags.resolution, "resolution", "r", "720p", "Resolution: 720p, 480p")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt
	prompt, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate duration
	if flags.duration < 1 || flags.duration > 15 {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 1 and 15 seconds")
	}

	// Validate aspect ratio
	if !validVideoAspects[flags.aspect] {
		return common.WriteError(cmd, "invalid_aspect", fmt.Sprintf("invalid aspect ratio '%s', use '16:9' or '9:16'", flags.aspect))
	}

	// Validate resolution
	if !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", fmt.Sprintf("invalid resolution '%s', use '720p' or '480p'", flags.resolution))
	}

	// Validate image if provided (before API key check for better error reporting)
	if flags.image != "" {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image file not found: %s", flags.image))
		}
		imgExt := strings.ToLower(filepath.Ext(flags.image))
		if _, ok := validImageFormats[imgExt]; !ok {
			return common.WriteError(cmd, "invalid_image_format", fmt.Sprintf("unsupported image format '%s', supported: png, jpeg, jpg", imgExt))
		}
	}

	// Check API key
	apiKey := config.GetAPIKey("XAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("XAI_API_KEY"))
	}

	// Check if image-to-video mode
	if flags.image != "" {
		return runCreateWithImage(cmd, prompt, flags, apiKey)
	}

	return runCreateTextOnly(cmd, prompt, flags, apiKey)
}

func runCreateTextOnly(cmd *cobra.Command, prompt string, flags *createFlags, apiKey string) error {
	// Build request body
	reqBody := map[string]any{
		"model":        "grok-2-video",
		"prompt":       prompt,
		"duration":     flags.duration,
		"aspect_ratio": flags.aspect,
		"resolution":   flags.resolution,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "json_error", err.Error())
	}

	// Make request
	req, err := http.NewRequest("POST", xaiBaseURL+videoGenerationsPath, bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return handleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	return parseCreateResponse(cmd, resp)
}

func runCreateWithImage(cmd *cobra.Command, prompt string, flags *createFlags, apiKey string) error {
	// Validate image exists
	if _, err := os.Stat(flags.image); os.IsNotExist(err) {
		return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image file not found: %s", flags.image))
	}

	// Validate image format
	imgExt := strings.ToLower(filepath.Ext(flags.image))
	if _, ok := validImageFormats[imgExt]; !ok {
		return common.WriteError(cmd, "invalid_image_format", fmt.Sprintf("unsupported image format '%s', supported: png, jpeg, jpg", imgExt))
	}

	// Read image file
	imgData, err := os.ReadFile(flags.image)
	if err != nil {
		return common.WriteError(cmd, "image_not_found", fmt.Sprintf("cannot read image file: %s", err.Error()))
	}

	// Build multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add image file
	imgFilename := filepath.Base(flags.image)
	imgPart, err := writer.CreateFormFile("image", imgFilename)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	imgPart.Write(imgData)

	// Add other fields
	writer.WriteField("model", "grok-2-video")
	writer.WriteField("prompt", prompt)
	writer.WriteField("duration", fmt.Sprintf("%d", flags.duration))
	writer.WriteField("aspect_ratio", flags.aspect)
	writer.WriteField("resolution", flags.resolution)

	writer.Close()

	// Make request
	req, err := http.NewRequest("POST", xaiBaseURL+videoGenerationsPath, &buf)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return handleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	return parseCreateResponse(cmd, resp)
}

func parseCreateResponse(cmd *cobra.Command, resp *http.Response) error {
	var apiResp xaiVideoCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for API error
	if apiResp.Error != nil {
		return handleAPIError(cmd, resp.StatusCode, apiResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d", resp.StatusCode))
	}

	status := apiResp.Status
	if status == "" {
		status = "pending"
	}

	return common.WriteSuccess(cmd, createResponse{
		Success:   true,
		RequestID: apiResp.RequestID,
		Status:    status,
	})
}

func getPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	// From positional argument
	if len(args) > 0 {
		prompt := strings.TrimSpace(strings.Join(args, " "))
		if prompt != "" {
			return prompt, nil
		}
	}

	// From file
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %w", err)
		}
		prompt := strings.TrimSpace(string(data))
		if prompt != "" {
			return prompt, nil
		}
	}

	// From stdin (only if not a terminal)
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", errors.New("no prompt provided, use positional argument, --prompt-file flag, or pipe from stdin")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %w", err)
		}
		prompt := strings.TrimSpace(string(data))
		if prompt != "" {
			return prompt, nil
		}
	}

	return "", errors.New("no prompt provided, use positional argument, --prompt-file flag, or pipe from stdin")
}

// encodeImageToBase64 encodes an image file to base64 with data URL prefix
func encodeImageToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType := validImageFormats[ext]

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}
