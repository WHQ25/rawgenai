package image

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

type createFlags struct {
	model      string
	ratio      string
	format     string
	imageRef   string
	styleRef   string
	modifyRef  string
	promptFile string
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:   "create <prompt>",
		Short: "Generate an image",
		Long: `Generate an image from a text prompt.

Optionally use reference images for style, content, or modification.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVarP(&flags.model, "model", "m", "photon-1", "Model (photon-1, photon-flash-1)")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "16:9", "Aspect ratio (1:1, 16:9, 9:16, 4:3, 3:4, 21:9, 9:21)")
	cmd.Flags().StringVar(&flags.format, "format", "jpg", "Output format (jpg, png)")
	cmd.Flags().StringVar(&flags.imageRef, "image-ref", "", "Image reference URL for content guidance")
	cmd.Flags().StringVar(&flags.styleRef, "style-ref", "", "Style reference URL")
	cmd.Flags().StringVar(&flags.modifyRef, "modify-ref", "", "Modify image reference URL")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Get prompt
	prompt, err := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", "prompt is required")
	}

	// Validate model
	if !validImageModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "model must be photon-1 or photon-flash-1")
	}

	// Validate ratio
	if !validAspectRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "ratio must be 1:1, 16:9, 9:16, 4:3, 3:4, 21:9, or 9:21")
	}

	// Validate format
	if !validImageFormats[flags.format] {
		return common.WriteError(cmd, "invalid_format", "format must be jpg or png")
	}

	// Validate image references (local files must exist)
	if flags.imageRef != "" && !shared.IsURL(flags.imageRef) {
		if _, err := os.Stat(flags.imageRef); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "image reference not found: "+flags.imageRef)
		}
	}

	if flags.styleRef != "" && !shared.IsURL(flags.styleRef) {
		if _, err := os.Stat(flags.styleRef); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "style reference not found: "+flags.styleRef)
		}
	}

	if flags.modifyRef != "" && !shared.IsURL(flags.modifyRef) {
		if _, err := os.Stat(flags.modifyRef); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "modify reference not found: "+flags.modifyRef)
		}
	}

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Build request body
	body := map[string]interface{}{
		"generation_type": "image",
		"model":           flags.model,
		"prompt":          prompt,
		"aspect_ratio":    flags.ratio,
		"format":          flags.format,
	}

	// Add image reference if provided
	if flags.imageRef != "" {
		imageURL, err := shared.ResolveImageURL(flags.imageRef)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", err.Error())
		}
		body["image_ref"] = []map[string]interface{}{
			{"url": imageURL},
		}
	}

	// Add style reference if provided
	if flags.styleRef != "" {
		styleURL, err := shared.ResolveImageURL(flags.styleRef)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", err.Error())
		}
		body["style_ref"] = []map[string]interface{}{
			{"url": styleURL},
		}
	}

	// Add modify reference if provided
	if flags.modifyRef != "" {
		modifyURL, err := shared.ResolveImageURL(flags.modifyRef)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", err.Error())
		}
		body["modify_image_ref"] = map[string]interface{}{
			"url": modifyURL,
		}
	}

	jsonBody, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/generations/image", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return shared.HandleAPIError(cmd, resp)
	}

	var gen shared.Generation
	if err := json.NewDecoder(resp.Body).Decode(&gen); err != nil {
		return common.WriteError(cmd, "decode_error", err.Error())
	}

	return common.WriteSuccess(cmd, map[string]interface{}{
		"task_id":    gen.ID,
		"state":      gen.State,
		"model":      gen.Model,
		"created_at": gen.CreatedAt,
	})
}
