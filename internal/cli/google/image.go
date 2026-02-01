package google

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

// Model name mapping
var modelIDs = map[string]string{
	"flash": "gemini-2.5-flash-image",
	"pro":   "gemini-3-pro-image-preview",
}

// Valid aspect ratios
var validAspects = map[string]bool{
	"1:1":  true,
	"2:3":  true,
	"3:2":  true,
	"3:4":  true,
	"4:3":  true,
	"4:5":  true,
	"5:4":  true,
	"9:16": true,
	"16:9": true,
	"21:9": true,
}

// Valid sizes (Pro model only)
var validSizes = map[string]bool{
	"1K": true,
	"2K": true,
	"4K": true,
}

// Max images per model
var maxImages = map[string]int{
	"flash": 3,
	"pro":   14,
}

// Response type
type imageResponse struct {
	Success bool   `json:"success"`
	File    string `json:"file,omitempty"`
	Model   string `json:"model,omitempty"`
	Aspect  string `json:"aspect,omitempty"`
	Size    string `json:"size,omitempty"`
}

// Flag struct
type imageFlags struct {
	output string
	images []string
	file   string
	model  string
	aspect string
	size   string
	search bool
}

// Command
var imageCmd = newImageCmd()

func newImageCmd() *cobra.Command {
	flags := &imageFlags{}

	cmd := &cobra.Command{
		Use:           "image [prompt]",
		Short:         "Generate and edit images using Google Gemini Nano Banana models",
		Long:          "Generate and edit images using Google Gemini Nano Banana models (flash and pro).",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImage(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (.png)")
	cmd.Flags().StringArrayVarP(&flags.images, "image", "i", nil, "Reference image(s), can be repeated")
	cmd.Flags().StringVar(&flags.file, "file", "", "Input prompt file")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "flash", "Model: flash, pro")
	cmd.Flags().StringVarP(&flags.aspect, "aspect", "a", "1:1", "Aspect ratio")
	cmd.Flags().StringVarP(&flags.size, "size", "s", "1K", "Image size (Pro only): 1K, 2K, 4K")
	cmd.Flags().BoolVar(&flags.search, "search", false, "Enable Google Search grounding (Pro only)")

	return cmd
}

func runImage(cmd *cobra.Command, args []string, flags *imageFlags) error {
	// Get prompt
	prompt, err := getPrompt(args, flags.file, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate format (only PNG supported)
	ext := strings.ToLower(filepath.Ext(flags.output))
	if ext != ".png" {
		return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s', only .png is supported", ext))
	}

	// Validate model
	modelID, ok := modelIDs[flags.model]
	if !ok {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s', use 'flash' or 'pro'", flags.model))
	}

	// Validate aspect ratio
	if !validAspects[flags.aspect] {
		validList := make([]string, 0, len(validAspects))
		for k := range validAspects {
			validList = append(validList, k)
		}
		return common.WriteError(cmd, "invalid_aspect", fmt.Sprintf("invalid aspect ratio '%s', valid: %s", flags.aspect, strings.Join(validList, ", ")))
	}

	// Validate size
	if !validSizes[flags.size] {
		return common.WriteError(cmd, "invalid_size", fmt.Sprintf("invalid size '%s', use '1K', '2K', or '4K' (uppercase K required)", flags.size))
	}

	// Size only supported in Pro model
	if flags.size != "1K" && flags.model != "pro" {
		return common.WriteError(cmd, "size_requires_pro", "--size 2K/4K requires --model pro")
	}

	// Search only supported in Pro model
	if flags.search && flags.model != "pro" {
		return common.WriteError(cmd, "search_requires_pro", "--search requires --model pro")
	}

	// Validate image count
	maxImg := maxImages[flags.model]
	if len(flags.images) > maxImg {
		return common.WriteError(cmd, "too_many_images", fmt.Sprintf("maximum %d reference images allowed for model '%s'", maxImg, flags.model))
	}

	// Validate image files exist
	for _, img := range flags.images {
		if _, err := os.Stat(img); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image file not found: %s", img))
		}
	}

	// Check API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "GEMINI_API_KEY or GOOGLE_API_KEY environment variable is not set")
	}

	// Build content parts
	var parts []*genai.Part

	// Add text prompt
	parts = append(parts, genai.NewPartFromText(prompt))

	// Add reference images
	for _, imgPath := range flags.images {
		imgData, err := os.ReadFile(imgPath)
		if err != nil {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("cannot read image file: %s", err.Error()))
		}

		mimeType := getMimeType(imgPath)
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: mimeType,
				Data:     imgData,
			},
		})
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
		ResponseModalities: []string{"IMAGE"},
		ImageConfig: &genai.ImageConfig{
			AspectRatio: flags.aspect,
		},
	}

	// Add size for Pro model
	if flags.model == "pro" {
		config.ImageConfig.ImageSize = flags.size
	}

	// Add Google Search tool for Pro model
	if flags.search {
		config.Tools = []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		}
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

	// Extract image from response
	var imageBytes []byte
	if result.Candidates != nil && len(result.Candidates) > 0 {
		for _, part := range result.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				imageBytes = part.InlineData.Data
				break
			}
		}
	}

	if imageBytes == nil {
		return common.WriteError(cmd, "no_image", "no image generated in response")
	}

	// Save image
	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	if err := os.WriteFile(absPath, imageBytes, 0644); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	// Build response
	resp := imageResponse{
		Success: true,
		File:    absPath,
		Model:   modelID,
		Aspect:  flags.aspect,
	}
	if flags.model == "pro" {
		resp.Size = flags.size
	}

	return common.WriteSuccess(cmd, resp)
}

func getPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
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

func getMimeType(path string) string {
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
	default:
		return "application/octet-stream"
	}
}

func handleAPIError(cmd *cobra.Command, err error) error {
	errStr := err.Error()

	// Check for common error patterns
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "invalid") && strings.Contains(errStr, "key") {
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or revoked")
	}
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "permission") {
		return common.WriteError(cmd, "permission_denied", "API key lacks required permissions")
	}
	if strings.Contains(errStr, "429") {
		if strings.Contains(errStr, "quota") {
			return common.WriteError(cmd, "quota_exceeded", "API quota exhausted")
		}
		return common.WriteError(cmd, "rate_limit", "too many requests")
	}
	if strings.Contains(errStr, "safety") || strings.Contains(errStr, "policy") {
		return common.WriteError(cmd, "content_policy", "content violates safety policy")
	}
	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") {
		return common.WriteError(cmd, "connection_error", "cannot connect to Gemini API")
	}

	return common.WriteError(cmd, "api_error", err.Error())
}
