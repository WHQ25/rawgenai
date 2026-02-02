package seed

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

// Model name mapping
var seedModelIDs = map[string]string{
	"4.5": "doubao-seedream-4-5-251128",
	"4.0": "doubao-seedream-4-0-250828",
}

// Valid sizes
var validSeedSizes = map[string]bool{
	"2K": true,
	"4K": true,
}

// Size regex for WxH format
var sizeRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)

// Response type
type seedImageResponse struct {
	Success bool     `json:"success"`
	Files   []string `json:"files,omitempty"`
	File    string   `json:"file,omitempty"`
	Model   string   `json:"model,omitempty"`
	Size    string   `json:"size,omitempty"`
	Count   int      `json:"count,omitempty"`
}

// Flag struct
type seedImageFlags struct {
	output     string
	images     []string
	promptFile string
	model      string
	size       string
	count      int
	watermark  bool
}

// Command
var imageCmd = newSeedImageCmd()

func newSeedImageCmd() *cobra.Command {
	flags := &seedImageFlags{}

	cmd := &cobra.Command{
		Use:           "image [prompt]",
		Short:         "Generate and edit images using ByteDance Seedream models",
		Long:          "Generate and edit images using ByteDance Seedream models (4.5 and 4.0).",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeedImage(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (required, .jpg/.jpeg)")
	cmd.Flags().StringArrayVarP(&flags.images, "image", "i", nil, "Reference image(s), can be repeated (max 14)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "4.5", "Model: 4.5, 4.0")
	cmd.Flags().StringVarP(&flags.size, "size", "s", "2K", "Image size: 2K, 4K, or WxH (e.g., 2048x2048)")
	cmd.Flags().IntVarP(&flags.count, "count", "n", 1, "Number of images to generate (1-10)")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Add watermark to output")

	return cmd
}

func runSeedImage(cmd *cobra.Command, args []string, flags *seedImageFlags) error {
	// Get prompt
	prompt, err := getSeedPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate format (only JPEG supported)
	ext := strings.ToLower(filepath.Ext(flags.output))
	if ext != ".jpg" && ext != ".jpeg" {
		return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s', only .jpg/.jpeg is supported", ext))
	}

	// Validate model
	modelID, ok := seedModelIDs[flags.model]
	if !ok {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s', use '4.5' or '4.0'", flags.model))
	}

	// Validate size
	if !isValidSeedSize(flags.size) {
		return common.WriteError(cmd, "invalid_size", fmt.Sprintf("invalid size '%s', use '2K', '4K', or 'WxH' format (e.g., 2048x2048)", flags.size))
	}

	// Validate count
	if flags.count < 1 || flags.count > 10 {
		return common.WriteError(cmd, "invalid_count", "count must be between 1 and 10")
	}

	// Validate image count
	if len(flags.images) > 14 {
		return common.WriteError(cmd, "too_many_images", "maximum 14 reference images allowed")
	}

	// Validate image files exist
	for _, img := range flags.images {
		if _, err := os.Stat(img); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image file not found: %s", img))
		}
	}

	// Check API key
	apiKey := config.GetAPIKey("ARK_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ARK_API_KEY"))
	}

	// Prepare image data
	var imageURLs []string
	for _, imgPath := range flags.images {
		imgData, err := os.ReadFile(imgPath)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image file: %s", err.Error()))
		}

		mimeType := getSeedMimeType(imgPath)
		base64Data := base64.StdEncoding.EncodeToString(imgData)
		dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
		imageURLs = append(imageURLs, dataURL)
	}

	// Build request body
	body := map[string]any{
		"model":           modelID,
		"prompt":          prompt,
		"size":            flags.size,
		"response_format": "b64_json",
		"watermark":       flags.watermark,
	}

	// Add images if provided
	if len(imageURLs) == 1 {
		body["image"] = imageURLs[0]
	} else if len(imageURLs) > 1 {
		body["image"] = imageURLs
	}

	// Add sequential generation options for multiple outputs
	if flags.count > 1 {
		body["sequential_image_generation"] = "auto"
		body["sequential_image_generation_options"] = map[string]any{
			"max_images": flags.count,
		}
	} else {
		body["sequential_image_generation"] = "disabled"
	}

	// Serialize request body
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://ark.cn-beijing.volces.com/api/v3/images/generations", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleSeedAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
			Size    string `json:"size"`
		} `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return common.WriteError(cmd, result.Error.Code, result.Error.Message)
		}
		return handleSeedHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	if result.Error != nil {
		return common.WriteError(cmd, result.Error.Code, result.Error.Message)
	}

	if len(result.Data) == 0 {
		return common.WriteError(cmd, "no_image", "no image generated in response")
	}

	// Save images
	var savedFiles []string
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	if len(result.Data) == 1 {
		// Single image
		imageBytes, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
		if err != nil {
			return common.WriteError(cmd, "decode_error", fmt.Sprintf("cannot decode image: %s", err.Error()))
		}

		if err := os.WriteFile(absPath, imageBytes, 0644); err != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
		}
		savedFiles = append(savedFiles, absPath)
	} else {
		// Multiple images - add index to filename
		baseName := strings.TrimSuffix(absPath, filepath.Ext(absPath))
		extName := filepath.Ext(absPath)

		for i, img := range result.Data {
			imageBytes, err := base64.StdEncoding.DecodeString(img.B64JSON)
			if err != nil {
				return common.WriteError(cmd, "decode_error", fmt.Sprintf("cannot decode image %d: %s", i+1, err.Error()))
			}

			outputPath := fmt.Sprintf("%s_%d%s", baseName, i+1, extName)
			if err := os.WriteFile(outputPath, imageBytes, 0644); err != nil {
				return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
			}
			savedFiles = append(savedFiles, outputPath)
		}
	}

	// Build response
	output := seedImageResponse{
		Success: true,
		Model:   modelID,
		Size:    flags.size,
		Count:   len(savedFiles),
	}

	if len(savedFiles) == 1 {
		output.File = savedFiles[0]
	} else {
		output.Files = savedFiles
	}

	return common.WriteSuccess(cmd, output)
}

func getSeedPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	// Priority 1: Positional argument
	if len(args) > 0 {
		text := strings.TrimSpace(args[0])
		if text != "" {
			return text, nil
		}
	}

	// Priority 2: File
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", fmt.Errorf("file is empty")
		}
		return text, nil
	}

	// Priority 3: Stdin (only if not a terminal)
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", fmt.Errorf("no prompt provided")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", fmt.Errorf("stdin is empty")
		}
		return text, nil
	}

	return "", fmt.Errorf("no prompt provided")
}

func isValidSeedSize(size string) bool {
	// Check preset sizes
	if validSeedSizes[size] {
		return true
	}

	// Check WxH format
	matches := sizeRegex.FindStringSubmatch(size)
	if matches == nil {
		return false
	}

	width, _ := strconv.Atoi(matches[1])
	height, _ := strconv.Atoi(matches[2])

	// Validate dimensions
	if width < 14 || height < 14 {
		return false
	}

	// Check aspect ratio (1/16 to 16)
	ratio := float64(width) / float64(height)
	if ratio < 1.0/16.0 || ratio > 16.0 {
		return false
	}

	// Check total pixels (921600 to 16777216 for 4.0, 3686400 to 16777216 for 4.5)
	// Use the more permissive range
	pixels := width * height
	if pixels < 921600 || pixels > 16777216 {
		return false
	}

	return true
}

func getSeedMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".tiff", ".tif":
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}

func handleSeedAPIError(cmd *cobra.Command, err error) error {
	errStr := err.Error()

	// Check for common error patterns
	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "refused") {
		return common.WriteError(cmd, "connection_error", "cannot connect to Ark API")
	}
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "dns") {
		return common.WriteError(cmd, "connection_error", "cannot resolve Ark API host")
	}

	return common.WriteError(cmd, "api_error", err.Error())
}

func handleSeedHTTPError(cmd *cobra.Command, statusCode int, body string) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
	case http.StatusForbidden:
		return common.WriteError(cmd, "permission_denied", "API key lacks required permissions")
	case http.StatusTooManyRequests:
		if strings.Contains(body, "quota") {
			return common.WriteError(cmd, "quota_exceeded", "API quota exhausted")
		}
		return common.WriteError(cmd, "rate_limit", "too many requests")
	case http.StatusBadRequest:
		if strings.Contains(body, "safety") || strings.Contains(body, "policy") {
			return common.WriteError(cmd, "content_policy", "content violates safety policy")
		}
		return common.WriteError(cmd, "invalid_request", "invalid request parameters")
	case http.StatusInternalServerError:
		return common.WriteError(cmd, "server_error", "Ark server error")
	case http.StatusServiceUnavailable:
		return common.WriteError(cmd, "server_overloaded", "Ark server overloaded")
	}

	return common.WriteError(cmd, "api_error", fmt.Sprintf("HTTP %d: %s", statusCode, body))
}
