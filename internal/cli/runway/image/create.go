package image

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var (
	validImageModels = map[string]bool{
		"gen4_image_turbo":  true,
		"gen4_image":        true,
		"gemini_2.5_flash":  true,
	}
	validImageRatios = map[string]bool{
		"1024:1024": true, "1080:1080": true, "1168:880": true,
		"1360:768": true, "1440:1080": true, "1080:1440": true,
		"1808:768": true, "1920:1080": true, "1080:1920": true,
		"2112:912": true, "1280:720": true, "720:1280": true,
		"720:720": true, "960:720": true, "720:960": true,
		"1680:720": true,
	}
)

type createFlags struct {
	refImages    []string
	refTags      []string
	model        string
	ratio        string
	seed         int
	promptFile   string
	publicFigure string
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create <prompt>",
		Short:         "Generate image from text/images",
		Long:          "Generate an image from text prompt and reference images using Runway AI.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringArrayVarP(&flags.refImages, "ref-image", "i", nil, "Reference image (1-3, can be specified multiple times)")
	cmd.Flags().StringArrayVar(&flags.refTags, "ref-tag", nil, "Tag for reference image (optional)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "gen4_image_turbo", "Model: gen4_image_turbo, gen4_image, gemini_2.5_flash")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "1024:1024", "Output resolution")
	cmd.Flags().IntVar(&flags.seed, "seed", -1, "Random seed (0-4294967295)")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")
	cmd.Flags().StringVar(&flags.publicFigure, "public-figure", "auto", "Content moderation: auto, low")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// 1. Get prompt (required)
	prompt, err := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", "prompt is required")
	}

	// 2. Validate required: ref-image
	if len(flags.refImages) == 0 {
		return common.WriteError(cmd, "missing_ref_image", "at least one reference image is required (--ref-image)")
	}

	// 3. Validate ref-image count
	if len(flags.refImages) > 3 {
		return common.WriteError(cmd, "too_many_ref_images", "maximum 3 reference images allowed")
	}

	// 4. Validate enum: model
	if !validImageModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "invalid model. Valid models: gen4_image_turbo, gen4_image, gemini_2.5_flash")
	}

	// 5. Validate enum: ratio
	if !validImageRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "invalid ratio")
	}

	// 6. Validate range: seed
	if flags.seed != -1 && (flags.seed < 0 || flags.seed > 4294967295) {
		return common.WriteError(cmd, "invalid_seed", "seed must be between 0 and 4294967295")
	}

	// 7. Validate enum: publicFigure
	if flags.publicFigure != "auto" && flags.publicFigure != "low" {
		return common.WriteError(cmd, "invalid_public_figure", "public-figure must be 'auto' or 'low'")
	}

	// 8. Validate file existence (local files only)
	for _, img := range flags.refImages {
		if !shared.IsURL(img) {
			if _, err := os.Stat(img); os.IsNotExist(err) {
				return common.WriteError(cmd, "image_not_found", "image file not found: "+img)
			}
		}
	}

	// 9. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 10. Resolve reference images
	refImages := make([]map[string]any, 0, len(flags.refImages))
	for i, img := range flags.refImages {
		imgURI, err := shared.ResolveMediaURI(img, "image")
		if err != nil {
			return common.WriteError(cmd, "image_read_error", "failed to read image: "+err.Error())
		}
		refImg := map[string]any{
			"uri": imgURI,
		}
		// Add tag if provided
		if i < len(flags.refTags) && strings.TrimSpace(flags.refTags[i]) != "" {
			refImg["tag"] = flags.refTags[i]
		}
		refImages = append(refImages, refImg)
	}

	// 11. Build request body
	body := map[string]any{
		"model":           flags.model,
		"promptText":      prompt,
		"ratio":           flags.ratio,
		"referenceImages": refImages,
	}
	if flags.seed >= 0 {
		body["seed"] = flags.seed
	}
	if flags.publicFigure != "auto" {
		body["contentModeration"] = map[string]string{
			"publicFigureThreshold": flags.publicFigure,
		}
	}

	// 12. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/text_to_image", bytes.NewReader(bodyJSON))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return shared.HandleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return shared.HandleAPIError(cmd, resp)
	}

	// 13. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 14. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
