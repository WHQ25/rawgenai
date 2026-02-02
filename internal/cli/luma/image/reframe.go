package image

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/luma/shared"
	"github.com/spf13/cobra"
)

type reframeFlags struct {
	image      string
	model      string
	ratio      string
	format     string
	promptFile string
}

func newReframeCmd() *cobra.Command {
	flags := &reframeFlags{}

	cmd := &cobra.Command{
		Use:   "reframe [prompt]",
		Short: "Reframe an image to a new aspect ratio",
		Long: `Reframe an image by changing its aspect ratio and filling in new content.

The AI will intelligently extend or crop the image to fit the new ratio.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReframe(cmd, args, flags)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Source image (URL or local file)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "photon-1", "Model (photon-1, photon-flash-1)")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "16:9", "Target aspect ratio")
	cmd.Flags().StringVar(&flags.format, "format", "jpg", "Output format (jpg, png)")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	cmd.MarkFlagRequired("image")

	return cmd
}

func runReframe(cmd *cobra.Command, args []string, flags *reframeFlags) error {
	// Validate image
	if flags.image == "" {
		return common.WriteError(cmd, "missing_image", "image is required (--image)")
	}

	// Validate local file exists
	if !shared.IsURL(flags.image) {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "image not found: "+flags.image)
		}
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

	// Get optional prompt
	prompt, _ := shared.GetPrompt(args, flags.promptFile, nil)

	// Check API key
	if shared.GetLumaAPIKey() == "" {
		return common.WriteError(cmd, "missing_api_key",
			"LUMA_API_KEY not found. Set it with: rawgenai config set luma_api_key <your-key>")
	}

	// Resolve image URL
	imageURL, err := shared.ResolveImageURL(flags.image)
	if err != nil {
		return common.WriteError(cmd, "image_read_error", err.Error())
	}

	// Build request body
	body := map[string]interface{}{
		"generation_type": "reframe_image",
		"model":           flags.model,
		"aspect_ratio":    flags.ratio,
		"format":          flags.format,
		"media": map[string]string{
			"url": imageURL,
		},
	}

	if prompt != "" {
		body["prompt"] = prompt
	}

	jsonBody, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/generations/image/reframe", bytes.NewReader(jsonBody))
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
