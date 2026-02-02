package grok

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

const (
	xaiBaseURL           = "https://api.x.ai/v1"
	imageGenerationsPath = "/images/generations"
	imageEditsPath       = "/images/edits"
)

// Valid aspect ratios for image generation
var validImageAspects = map[string]bool{
	"1:1":  true,
	"16:9": true,
	"9:16": true,
	"4:3":  true,
	"3:4":  true,
}

// Supported image formats
var supportedImageFormats = map[string]string{
	".png":  "image/png",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
}

type imageFlags struct {
	output     string
	promptFile string
	image      string
	n          int
	aspect     string
}

type imageResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

// API response types
type xaiImageResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
	} `json:"data"`
	Error *xaiError `json:"error,omitempty"`
}

type xaiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

var imageCmd = newImageCmd()

func newImageCmd() *cobra.Command {
	flags := &imageFlags{}

	cmd := &cobra.Command{
		Use:   "image [prompt]",
		Short: "Generate or edit images using xAI Grok",
		Long: `Generate or edit images using xAI Grok API.

Without --image: Generation mode (POST /v1/images/generations)
With --image: Edit mode (POST /v1/images/edits)`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImage(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (required)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read prompt from file")
	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Input image for edit mode")
	cmd.Flags().IntVarP(&flags.n, "n", "n", 1, "Number of images to generate (1-10, generation mode only)")
	cmd.Flags().StringVarP(&flags.aspect, "aspect", "a", "1:1", "Aspect ratio (generation mode only)")

	return cmd
}

func runImage(cmd *cobra.Command, args []string, flags *imageFlags) error {
	// Get prompt
	prompt, err := getPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate output format
	ext := strings.ToLower(filepath.Ext(flags.output))
	if _, ok := supportedImageFormats[ext]; !ok {
		return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s', supported: png, jpeg, jpg", ext))
	}

	// Determine mode based on --image flag
	isEditMode := flags.image != ""

	if isEditMode {
		return runImageEdit(cmd, prompt, flags)
	}
	return runImageGenerate(cmd, prompt, flags)
}

func runImageGenerate(cmd *cobra.Command, prompt string, flags *imageFlags) error {
	// Validate n
	if flags.n < 1 || flags.n > 10 {
		return common.WriteError(cmd, "invalid_n", "n must be between 1 and 10")
	}

	// Validate aspect ratio
	if !validImageAspects[flags.aspect] {
		return common.WriteError(cmd, "invalid_aspect", fmt.Sprintf("invalid aspect ratio '%s', valid: 1:1, 16:9, 9:16, 4:3, 3:4", flags.aspect))
	}

	// Check API key
	apiKey := config.GetAPIKey("XAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("XAI_API_KEY"))
	}

	// Build request body
	reqBody := map[string]any{
		"model":           "grok-2-image",
		"prompt":          prompt,
		"n":               flags.n,
		"response_format": "b64_json",
	}

	// Note: aspect ratio may need different field name based on actual API
	// Using standard field names, adjust if API differs
	if flags.aspect != "1:1" {
		reqBody["aspect_ratio"] = flags.aspect
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return common.WriteError(cmd, "json_error", err.Error())
	}

	// Make request
	req, err := http.NewRequest("POST", xaiBaseURL+imageGenerationsPath, bytes.NewReader(jsonBody))
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

	// Parse response
	var apiResp xaiImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for API error
	if apiResp.Error != nil {
		return handleXAIError(cmd, resp.StatusCode, apiResp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d", resp.StatusCode))
	}

	// Extract image
	if len(apiResp.Data) == 0 {
		return common.WriteError(cmd, "no_image", "no image generated in response")
	}

	// Decode and save image
	imgData, err := base64.StdEncoding.DecodeString(apiResp.Data[0].B64JSON)
	if err != nil {
		return common.WriteError(cmd, "decode_error", fmt.Sprintf("cannot decode image: %s", err.Error()))
	}

	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	if err := os.WriteFile(absPath, imgData, 0644); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	return common.WriteSuccess(cmd, imageResponse{
		Success: true,
		File:    absPath,
		Mode:    "generate",
	})
}

func runImageEdit(cmd *cobra.Command, prompt string, flags *imageFlags) error {
	// Validate image file exists
	if _, err := os.Stat(flags.image); os.IsNotExist(err) {
		return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image file not found: %s", flags.image))
	}

	// Validate image format
	imgExt := strings.ToLower(filepath.Ext(flags.image))
	mimeType, ok := supportedImageFormats[imgExt]
	if !ok {
		return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported image format '%s', supported: png, jpeg, jpg", imgExt))
	}

	// Check API key
	apiKey := config.GetAPIKey("XAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("XAI_API_KEY"))
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

	// Add prompt
	writer.WriteField("prompt", prompt)

	// Add model
	writer.WriteField("model", "grok-2-image")

	// Add response format
	writer.WriteField("response_format", "b64_json")

	writer.Close()

	// Make request
	req, err := http.NewRequest("POST", xaiBaseURL+imageEditsPath, &buf)
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

	// Parse response
	var apiResp xaiImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check for API error
	if apiResp.Error != nil {
		return handleXAIError(cmd, resp.StatusCode, apiResp.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d", resp.StatusCode))
	}

	// Extract image
	if len(apiResp.Data) == 0 {
		return common.WriteError(cmd, "no_image", "no image generated in response")
	}

	// Decode and save image
	resultData, err := base64.StdEncoding.DecodeString(apiResp.Data[0].B64JSON)
	if err != nil {
		return common.WriteError(cmd, "decode_error", fmt.Sprintf("cannot decode image: %s", err.Error()))
	}

	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	if err := os.WriteFile(absPath, resultData, 0644); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	// Suppress unused variable warning
	_ = mimeType

	return common.WriteSuccess(cmd, imageResponse{
		Success: true,
		File:    absPath,
		Mode:    "edit",
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

func handleHTTPError(cmd *cobra.Command, err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") {
		return common.WriteError(cmd, "connection_error", "cannot connect to xAI API")
	}
	return common.WriteError(cmd, "network_error", err.Error())
}

func handleXAIError(cmd *cobra.Command, statusCode int, xaiErr *xaiError) error {
	switch statusCode {
	case 400:
		return common.WriteError(cmd, "invalid_request", xaiErr.Message)
	case 401:
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
	case 403:
		return common.WriteError(cmd, "permission_denied", "API key lacks required permissions")
	case 429:
		if strings.Contains(xaiErr.Message, "quota") {
			return common.WriteError(cmd, "quota_exceeded", xaiErr.Message)
		}
		return common.WriteError(cmd, "rate_limit", xaiErr.Message)
	case 500:
		return common.WriteError(cmd, "server_error", "xAI server error")
	case 503:
		return common.WriteError(cmd, "server_overloaded", "xAI server overloaded")
	default:
		return common.WriteError(cmd, "api_error", xaiErr.Message)
	}
}
