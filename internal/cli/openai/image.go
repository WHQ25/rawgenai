package openai

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	oai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/spf13/cobra"
)

var supportedImageFormats = map[string]string{
	".png":  "png",
	".jpeg": "jpeg",
	".jpg":  "jpeg",
	".webp": "webp",
}

// Response type
type imageResponse struct {
	Success    bool   `json:"success"`
	File       string `json:"file,omitempty"`
	Model      string `json:"model,omitempty"`
	ResponseID string `json:"response_id,omitempty"`
}

// Flag struct
type imageFlags struct {
	output      string
	images      []string
	mask        string
	promptFile  string
	continueID  string
	model       string
	size        string
	quality     string
	background  string
	compression int
	fidelity    string
	moderation  string
}

// Command
var imageCmd = newImageCmd()

func newImageCmd() *cobra.Command {
	flags := &imageFlags{}

	cmd := &cobra.Command{
		Use:           "image [prompt]",
		Short:         "Generate and edit images using OpenAI Responses API",
		Long:          "Generate and edit images using OpenAI Responses API with GPT Image models. Supports multiple reference images.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImage(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path (format from extension)")
	cmd.Flags().StringArrayVarP(&flags.images, "image", "i", nil, "Reference image(s), can be repeated (max 16)")
	cmd.Flags().StringVar(&flags.mask, "mask", "", "Mask image for inpainting (PNG with alpha)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Input prompt file")
	cmd.Flags().StringVarP(&flags.continueID, "continue", "c", "", "Previous response ID for multi-turn conversation")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "gpt-image-1", "Model name")
	cmd.Flags().StringVarP(&flags.size, "size", "s", "auto", "Image dimensions")
	cmd.Flags().StringVarP(&flags.quality, "quality", "q", "auto", "Image quality")
	cmd.Flags().StringVar(&flags.background, "background", "auto", "Background type")
	cmd.Flags().IntVar(&flags.compression, "compression", 100, "Compression 0-100 (JPEG/WebP only)")
	cmd.Flags().StringVar(&flags.fidelity, "fidelity", "low", "Input image fidelity (high/low)")
	cmd.Flags().StringVar(&flags.moderation, "moderation", "auto", "Moderation level")

	return cmd
}

func runImage(cmd *cobra.Command, args []string, flags *imageFlags) error {
	// Get prompt
	prompt, err := getText(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Validate output
	if flags.output == "" {
		return common.WriteError(cmd, "missing_output", "output file is required, use -o flag")
	}

	// Validate format
	ext := strings.ToLower(filepath.Ext(flags.output))
	outputFormat, ok := supportedImageFormats[ext]
	if !ok {
		return common.WriteError(cmd, "unsupported_format", fmt.Sprintf("unsupported format '%s', supported: png, jpeg, jpg, webp", ext))
	}

	// Validate compression
	if flags.compression < 0 || flags.compression > 100 {
		return common.WriteError(cmd, "invalid_compression", "compression must be between 0 and 100")
	}

	// Validate fidelity
	if flags.fidelity != "high" && flags.fidelity != "low" {
		return common.WriteError(cmd, "invalid_fidelity", "fidelity must be 'high' or 'low'")
	}

	// Validate transparent background requires png or webp
	if flags.background == "transparent" && ext != ".png" && ext != ".webp" {
		return common.WriteError(cmd, "transparent_requires_png_webp", "--background=transparent requires .png or .webp output")
	}

	// Validate mask requires image
	if flags.mask != "" && len(flags.images) == 0 {
		return common.WriteError(cmd, "mask_requires_image", "--mask requires at least one --image")
	}

	// Validate image count
	if len(flags.images) > 16 {
		return common.WriteError(cmd, "too_many_images", "maximum 16 reference images allowed")
	}

	// Validate image files exist
	for _, img := range flags.images {
		if _, err := os.Stat(img); os.IsNotExist(err) {
			return common.WriteError(cmd, "file_not_found", fmt.Sprintf("image file not found: %s", img))
		}
	}

	// Validate mask file exists
	if flags.mask != "" {
		if _, err := os.Stat(flags.mask); os.IsNotExist(err) {
			return common.WriteError(cmd, "file_not_found", fmt.Sprintf("mask file not found: %s", flags.mask))
		}
	}

	// Check API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "OPENAI_API_KEY environment variable is not set")
	}

	// Build input content
	content := []responses.ResponseInputContentUnionParam{
		responses.ResponseInputContentParamOfInputText(prompt),
	}

	// Add reference images
	for _, imgPath := range flags.images {
		imgData, err := os.ReadFile(imgPath)
		if err != nil {
			return common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot read image file: %s", err.Error()))
		}

		mimeType := getMimeType(imgPath)
		base64Img := base64.StdEncoding.EncodeToString(imgData)
		dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Img)

		content = append(content, responses.ResponseInputContentUnionParam{
			OfInputImage: &responses.ResponseInputImageParam{
				ImageURL: oai.String(dataURL),
				Detail:   responses.ResponseInputImageDetailAuto,
			},
		})
	}

	// Build image generation tool
	tool := responses.ToolImageGenerationParam{
		Model:        flags.model,
		Size:         flags.size,
		Quality:      flags.quality,
		Background:   flags.background,
		OutputFormat: outputFormat,
		Moderation:   flags.moderation,
	}

	// Add compression for jpeg/webp
	if (outputFormat == "jpeg" || outputFormat == "webp") && flags.compression < 100 {
		tool.OutputCompression = oai.Int(int64(flags.compression))
	}

	// Add fidelity if images are provided
	if len(flags.images) > 0 {
		tool.InputFidelity = flags.fidelity
	}

	// Add mask if provided
	if flags.mask != "" {
		maskData, err := os.ReadFile(flags.mask)
		if err != nil {
			return common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot read mask file: %s", err.Error()))
		}

		mimeType := getMimeType(flags.mask)
		base64Mask := base64.StdEncoding.EncodeToString(maskData)
		maskURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Mask)

		tool.InputImageMask = responses.ToolImageGenerationInputImageMaskParam{
			ImageURL: oai.String(maskURL),
		}
	}

	// Build request params
	params := responses.ResponseNewParams{
		Model: oai.ResponsesModel("gpt-4.1"),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				{
					OfMessage: &responses.EasyInputMessageParam{
						Role: responses.EasyInputMessageRoleUser,
						Content: responses.EasyInputMessageContentUnionParam{
							OfInputItemContentList: content,
						},
					},
				},
			},
		},
		Tools: []responses.ToolUnionParam{
			{OfImageGeneration: &tool},
		},
	}

	// Add previous response ID for multi-turn conversation
	if flags.continueID != "" {
		params.PreviousResponseID = oai.String(flags.continueID)
	}

	// Call API
	client := oai.NewClient()
	ctx := context.Background()

	resp, err := client.Responses.New(ctx, params)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	// Extract image from response
	var imageBase64 string
	for _, output := range resp.Output {
		if output.Type == "image_generation_call" {
			call := output.AsImageGenerationCall()
			imageBase64 = call.Result
			break
		}
	}

	if imageBase64 == "" {
		return common.WriteError(cmd, "no_image", "no image generated in response")
	}

	// Decode and save image
	imgData, err := base64.StdEncoding.DecodeString(imageBase64)
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
		Success:    true,
		File:       absPath,
		Model:      flags.model,
		ResponseID: resp.ID,
	})
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
