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

var validV2VRatios = map[string]bool{
	"1280:720": true,
	"720:1280": true,
	"1104:832": true,
	"960:960":  true,
	"832:1104": true,
	"1584:672": true,
	"848:480":  true,
	"640:480":  true,
}

type video2videoFlags struct {
	video        string
	ratio        string
	seed         int
	refImage     string
	promptFile   string
	publicFigure string
}

func newVideo2VideoCmd() *cobra.Command {
	flags := &video2videoFlags{}

	cmd := &cobra.Command{
		Use:           "video2video <prompt>",
		Short:         "Generate video from video",
		Long:          "Generate a new video from an input video using Runway AI models.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideo2Video(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.video, "video", "v", "", "Input video (URL or local path)")
	cmd.Flags().StringVarP(&flags.ratio, "ratio", "r", "1280:720", "Output resolution")
	cmd.Flags().IntVar(&flags.seed, "seed", -1, "Random seed (0-4294967295)")
	cmd.Flags().StringVar(&flags.refImage, "ref-image", "", "Reference image for style")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")
	cmd.Flags().StringVar(&flags.publicFigure, "public-figure", "auto", "Content moderation: auto, low")

	return cmd
}

func runVideo2Video(cmd *cobra.Command, args []string, flags *video2videoFlags) error {
	// 1. Get prompt (required)
	prompt, err := shared.GetPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", "prompt is required")
	}

	// 2. Validate required: video
	if flags.video == "" {
		return common.WriteError(cmd, "missing_video", "input video is required (-v)")
	}

	// 3. Validate enum: ratio
	if !validV2VRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "invalid ratio. Valid ratios: 1280:720, 720:1280, 1104:832, 960:960, 832:1104, 1584:672, 848:480, 640:480")
	}

	// 4. Validate range: seed
	if flags.seed != -1 && (flags.seed < 0 || flags.seed > 4294967295) {
		return common.WriteError(cmd, "invalid_seed", "seed must be between 0 and 4294967295")
	}

	// 5. Validate enum: publicFigure
	if flags.publicFigure != "auto" && flags.publicFigure != "low" {
		return common.WriteError(cmd, "invalid_public_figure", "public-figure must be 'auto' or 'low'")
	}

	// 6. Validate file existence (local files only)
	if !shared.IsURL(flags.video) {
		if _, err := os.Stat(flags.video); os.IsNotExist(err) {
			return common.WriteError(cmd, "video_not_found", "video file not found: "+flags.video)
		}
	}
	if flags.refImage != "" && !shared.IsURL(flags.refImage) {
		if _, err := os.Stat(flags.refImage); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", "reference image file not found: "+flags.refImage)
		}
	}

	// 7. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 8. Resolve video URI
	videoURI, err := shared.ResolveMediaURI(flags.video, "video")
	if err != nil {
		return common.WriteError(cmd, "video_read_error", "failed to read video: "+err.Error())
	}

	// 9. Build request body
	body := map[string]any{
		"model":      "gen4_aleph",
		"videoUri":   videoURI,
		"promptText": prompt,
		"ratio":      flags.ratio,
	}
	if flags.seed >= 0 {
		body["seed"] = flags.seed
	}
	if flags.refImage != "" {
		refImageURI, err := shared.ResolveMediaURI(flags.refImage, "image")
		if err != nil {
			return common.WriteError(cmd, "image_read_error", "failed to read reference image: "+err.Error())
		}
		body["references"] = []map[string]any{
			{
				"type": "image",
				"uri":  refImageURI,
			},
		}
	}
	if flags.publicFigure != "auto" {
		body["contentModeration"] = map[string]string{
			"publicFigureThreshold": flags.publicFigure,
		}
	}

	// 10. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/video_to_video", bytes.NewReader(bodyJSON))
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

	// 11. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 12. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
