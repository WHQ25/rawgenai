package dashscope

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
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

const (
	imageSyncPath = "/services/aigc/multimodal-generation/generation"
)

// Text-to-image models
var validT2IModels = map[string]bool{
	"wan2.6-t2i":     true,
	"qwen-image-max":  true,
	"qwen-image-plus": true,
}

// Image editing models
var validEditModels = map[string]bool{
	"qwen-image-edit-max":  true,
	"qwen-image-edit-plus": true,
	"qwen-image-edit":      true,
	"wan2.6-image":         true,
}

// Max input images per edit model
var maxInputImages = map[string]int{
	"qwen-image-edit-max":  3,
	"qwen-image-edit-plus": 3,
	"qwen-image-edit":      3,
	"wan2.6-image":         4,
}

// Max output count per model
var maxOutputCount = map[string]int{
	"wan2.6-t2i":           4,
	"qwen-image-max":       1,
	"qwen-image-plus":      1,
	"qwen-image-edit-max":  6,
	"qwen-image-edit-plus": 6,
	"qwen-image-edit":      1,
	"wan2.6-image":         4,
}

// Models that do NOT support prompt_extend
var noPromptExtendModels = map[string]bool{
	"qwen-image-edit": true,
}

type imageFlags struct {
	output       string
	images       []string
	promptFile   string
	model        string
	size         string
	count        int
	negative     string
	seed         int
	promptExtend bool
	watermark    bool
}

type imageResponse struct {
	Success bool          `json:"success"`
	Model   string        `json:"model"`
	File    string        `json:"file,omitempty"`
	Files   []string      `json:"files,omitempty"`
	Images  []imageResult `json:"images,omitempty"`
}

type imageResult struct {
	URL   string `json:"url"`
	Index int    `json:"index"`
}

// Command
var imageCmd = newImageCmd()

func newImageCmd() *cobra.Command {
	flags := &imageFlags{}

	cmd := &cobra.Command{
		Use:   "image [prompt]",
		Short: "Generate or edit images",
		Long: `Generate or edit images using DashScope models.

Without --image: Text-to-image (default model: wan2.6-t2i)
With --image: Image editing (default model: qwen-image-edit-plus)`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImage(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (auto-download)")
	cmd.Flags().StringArrayVarP(&flags.images, "image", "i", nil, "Input image(s) for editing (repeatable, local path or URL)")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "", "Model name (auto-selected by input)")
	cmd.Flags().StringVarP(&flags.size, "size", "s", "", "Image size \"width*height\"")
	cmd.Flags().IntVarP(&flags.count, "count", "n", 1, "Number of output images")
	cmd.Flags().StringVar(&flags.negative, "negative", "", "Negative prompt (max 500 chars)")
	cmd.Flags().IntVar(&flags.seed, "seed", 0, "Random seed [0, 2147483647]")
	cmd.Flags().BoolVar(&flags.promptExtend, "prompt-extend", true, "AI prompt enhancement")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Add AI-generated watermark")

	return cmd
}

func runImage(cmd *cobra.Command, args []string, flags *imageFlags) error {
	// 1. Get prompt
	prompt, err := getVideoPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	hasImages := len(flags.images) > 0

	// 2. Auto-select model if not specified
	model := flags.model
	if model == "" {
		if hasImages {
			model = "qwen-image-edit-plus"
		} else {
			model = "wan2.6-t2i"
		}
	}

	// 3. Validate model
	isT2I := validT2IModels[model]
	isEdit := validEditModels[model]
	if !isT2I && !isEdit {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("unknown model '%s'", model))
	}

	// 4. Image + model compatibility
	if hasImages && isT2I {
		return common.WriteError(cmd, "incompatible_image", fmt.Sprintf("--image is not supported by text-to-image model '%s'", model))
	}
	if !hasImages && isEdit {
		return common.WriteError(cmd, "missing_image", fmt.Sprintf("model '%s' requires --image input", model))
	}

	// 5. Image count limit
	if hasImages {
		maxImages := maxInputImages[model]
		if len(flags.images) > maxImages {
			return common.WriteError(cmd, "too_many_images", fmt.Sprintf("model '%s' accepts at most %d input images, got %d", model, maxImages, len(flags.images)))
		}
	}

	// 6. Image file existence (skip URLs)
	if hasImages {
		for _, img := range flags.images {
			if !isURL(img) {
				if _, err := os.Stat(img); os.IsNotExist(err) {
					return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image file not found: %s", img))
				}
			}
		}
	}

	// 7. Count validation
	if flags.count < 1 {
		return common.WriteError(cmd, "invalid_count", "count must be >= 1")
	}
	maxCount := maxOutputCount[model]
	if flags.count > maxCount {
		return common.WriteError(cmd, "invalid_count", fmt.Sprintf("model '%s' supports at most %d output images, got %d", model, maxCount, flags.count))
	}

	// 8. Compatibility: prompt_extend
	if noPromptExtendModels[model] && cmd.Flags().Changed("prompt-extend") && flags.promptExtend {
		return common.WriteError(cmd, "incompatible_prompt_extend", fmt.Sprintf("prompt-extend is not supported by model '%s'", model))
	}

	// 9. API key
	apiKey := config.GetAPIKey("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("DASHSCOPE_API_KEY"))
	}

	// Build request body
	content := []map[string]any{}

	// Add images first (for edit models)
	if hasImages {
		for _, img := range flags.images {
			imageValue, err := resolveImageInput(img)
			if err != nil {
				return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image: %s", err.Error()))
			}
			content = append(content, map[string]any{"image": imageValue})
		}
	}

	// Add text prompt
	content = append(content, map[string]any{"text": prompt})

	body := map[string]any{
		"model": model,
		"input": map[string]any{
			"messages": []map[string]any{
				{
					"role":    "user",
					"content": content,
				},
			},
		},
	}

	// Parameters
	params := map[string]any{
		"n": flags.count,
	}
	if flags.size != "" {
		params["size"] = flags.size
	}
	if flags.negative != "" {
		params["negative_prompt"] = flags.negative
	}
	if flags.seed != 0 {
		params["seed"] = flags.seed
	}
	if !noPromptExtendModels[model] {
		params["prompt_extend"] = flags.promptExtend
	}
	if flags.watermark {
		params["watermark"] = true
	}
	if len(params) > 0 {
		body["parameters"] = params
	}

	// Send request
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	reqURL := getBaseURL() + imageSyncPath
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return handleHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	// Parse response
	var result struct {
		Output struct {
			Choices []struct {
				Message struct {
					Content []struct {
						Image string `json:"image,omitempty"`
					} `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
		RequestID string `json:"request_id"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Extract image URLs
	var imageURLs []string
	for _, choice := range result.Output.Choices {
		for _, c := range choice.Message.Content {
			if c.Image != "" {
				imageURLs = append(imageURLs, c.Image)
			}
		}
	}

	if len(imageURLs) == 0 {
		return common.WriteError(cmd, "api_error", "no images in response")
	}

	// Build image results
	images := make([]imageResult, len(imageURLs))
	for i, url := range imageURLs {
		images[i] = imageResult{URL: url, Index: i}
	}

	output := imageResponse{
		Success: true,
		Model:   model,
		Images:  images,
	}

	// Download if -o is set
	if flags.output != "" {
		absPath, err := filepath.Abs(flags.output)
		if err != nil {
			absPath = flags.output
		}

		// Create output directory if needed
		dir := filepath.Dir(absPath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create directory: %s", err.Error()))
			}
		}

		if len(imageURLs) == 1 {
			if err := downloadFile(imageURLs[0], absPath); err != nil {
				return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download image: %s", err.Error()))
			}
			output.File = absPath
		} else {
			baseName := strings.TrimSuffix(absPath, filepath.Ext(absPath))
			extName := filepath.Ext(absPath)
			var files []string
			for i, url := range imageURLs {
				outputPath := fmt.Sprintf("%s_%d%s", baseName, i, extName)
				if err := downloadFile(url, outputPath); err != nil {
					return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download image %d: %s", i, err.Error()))
				}
				files = append(files, outputPath)
			}
			output.Files = files
		}
	}

	return common.WriteSuccess(cmd, output)
}

// resolveImageInput returns a URL as-is, or encodes a local file as base64 data URI.
func resolveImageInput(input string) (string, error) {
	if isURL(input) {
		return input, nil
	}

	data, err := os.ReadFile(input)
	if err != nil {
		return "", err
	}

	mime := imageMimeType(input)
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

func imageMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	case ".tiff", ".tif":
		return "image/tiff"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func downloadFile(url, outputPath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}
