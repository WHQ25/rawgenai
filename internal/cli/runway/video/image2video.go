package video

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/runway/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var (
	validI2VModels = map[string]bool{
		"gen4_turbo":  true,
		"veo3.1":      true,
		"veo3.1_fast": true,
		"gen3a_turbo": true,
		"veo3":        true,
	}
	validI2VRatios = map[string]bool{
		"1280:720": true,
		"720:1280": true,
		"1104:832": true,
		"832:1104": true,
		"960:960":  true,
		"1584:672": true,
	}
)

type image2videoFlags struct {
	image        string
	model        string
	ratio        string
	duration     int
	seed         int
	promptFile   string
	publicFigure string
}

func newImage2VideoCmd() *cobra.Command {
	flags := &image2videoFlags{}

	cmd := &cobra.Command{
		Use:           "image2video [prompt]",
		Short:         "Generate video from image",
		Long:          "Generate a video from an input image using Runway AI models.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImage2Video(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Input image (URL or local path)")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "gen4_turbo", "Model: gen4_turbo, veo3.1, veo3.1_fast, gen3a_turbo, veo3")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "1280:720", "Output resolution")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 5, "Duration in seconds (2-10)")
	cmd.Flags().IntVar(&flags.seed, "seed", -1, "Random seed (0-4294967295)")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")
	cmd.Flags().StringVar(&flags.publicFigure, "public-figure", "auto", "Content moderation: auto, low")

	return cmd
}

func runImage2Video(cmd *cobra.Command, args []string, flags *image2videoFlags) error {
	// 1. Validate required: image
	if flags.image == "" {
		return common.WriteError(cmd, "missing_image", "input image is required (-i)")
	}

	// 2. Validate enum: model
	if !validI2VModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", "invalid model. Valid models: gen4_turbo, veo3.1, veo3.1_fast, gen3a_turbo, veo3")
	}

	// 3. Validate enum: ratio
	if !validI2VRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "invalid ratio. Valid ratios: 1280:720, 720:1280, 1104:832, 832:1104, 960:960, 1584:672")
	}

	// 4. Validate range: duration
	if flags.duration < 2 || flags.duration > 10 {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 2 and 10 seconds")
	}

	// 5. Validate range: seed
	if flags.seed != -1 && (flags.seed < 0 || flags.seed > 4294967295) {
		return common.WriteError(cmd, "invalid_seed", "seed must be between 0 and 4294967295")
	}

	// 6. Validate enum: publicFigure
	if flags.publicFigure != "auto" && flags.publicFigure != "low" {
		return common.WriteError(cmd, "invalid_public_figure", "public-figure must be 'auto' or 'low'")
	}

	// 7. Validate file existence (local files only)
	if !shared.IsURL(flags.image) {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "image file not found: "+flags.image)
		}
	}

	// 8. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 9. Get prompt (optional for image2video)
	prompt, _ := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())

	// 10. Resolve image URI
	imageURI, err := shared.ResolveMediaURI(flags.image, "image")
	if err != nil {
		return common.WriteError(cmd, "image_read_error", "failed to read image: "+err.Error())
	}

	// 11. Build request body
	body := map[string]any{
		"model":       flags.model,
		"promptImage": imageURI,
		"ratio":       flags.ratio,
		"duration":    flags.duration,
	}
	if prompt != "" {
		body["promptText"] = prompt
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
	req, err := shared.CreateRequest("POST", "/v1/image_to_video", bytes.NewReader(bodyJSON))
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
