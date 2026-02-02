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

var validCharacterRatios = map[string]bool{
	"1280:720": true,
	"720:1280": true,
	"960:960":  true,
	"1104:832": true,
	"832:1104": true,
	"1584:672": true,
}

type characterFlags struct {
	character     string
	characterType string
	reference     string
	seed          int
	bodyControl   bool
	expression    int
	ratio         string
	publicFigure  string
}

func newCharacterCmd() *cobra.Command {
	flags := &characterFlags{}

	cmd := &cobra.Command{
		Use:           "character",
		Short:         "Control character performance",
		Long:          "Control a character's facial expressions and body movements using a reference video (Act Two).",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCharacter(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.character, "character", "c", "", "Character image or video")
	cmd.Flags().StringVar(&flags.characterType, "character-type", "image", "Character type: image, video")
	cmd.Flags().StringVarP(&flags.reference, "reference", "r", "", "Reference performance video")
	cmd.Flags().IntVar(&flags.seed, "seed", -1, "Random seed (0-4294967295)")
	cmd.Flags().BoolVar(&flags.bodyControl, "body-control", false, "Enable body control")
	cmd.Flags().IntVarP(&flags.expression, "expression", "e", 3, "Expression intensity (1-5)")
	cmd.Flags().StringVar(&flags.ratio, "ratio", "1280:720", "Output resolution")
	cmd.Flags().StringVar(&flags.publicFigure, "public-figure", "auto", "Content moderation: auto, low")

	return cmd
}

func runCharacter(cmd *cobra.Command, args []string, flags *characterFlags) error {
	// 1. Validate required: character
	if flags.character == "" {
		return common.WriteError(cmd, "missing_character", "character is required (-c)")
	}

	// 2. Validate required: reference
	if flags.reference == "" {
		return common.WriteError(cmd, "missing_reference", "reference video is required (-r)")
	}

	// 3. Validate enum: characterType
	if flags.characterType != "image" && flags.characterType != "video" {
		return common.WriteError(cmd, "invalid_character_type", "character-type must be 'image' or 'video'")
	}

	// 4. Validate range: expression
	if flags.expression < 1 || flags.expression > 5 {
		return common.WriteError(cmd, "invalid_expression", "expression must be between 1 and 5")
	}

	// 5. Validate enum: ratio
	if !validCharacterRatios[flags.ratio] {
		return common.WriteError(cmd, "invalid_ratio", "invalid ratio. Valid ratios: 1280:720, 720:1280, 960:960, 1104:832, 832:1104, 1584:672")
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
	if !shared.IsURL(flags.character) {
		if _, err := os.Stat(flags.character); os.IsNotExist(err) {
			return common.WriteError(cmd, "character_not_found", "character file not found: "+flags.character)
		}
	}
	if !shared.IsURL(flags.reference) {
		if _, err := os.Stat(flags.reference); os.IsNotExist(err) {
			return common.WriteError(cmd, "reference_not_found", "reference video not found: "+flags.reference)
		}
	}

	// 9. Check API key
	apiKey := shared.GetRunwayAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("RUNWAY_API_KEY"))
	}

	// 10. Resolve character URI
	var characterURI string
	var err error
	if flags.characterType == "image" {
		characterURI, err = shared.ResolveMediaURI(flags.character, "image")
	} else {
		characterURI, err = shared.ResolveMediaURI(flags.character, "video")
	}
	if err != nil {
		return common.WriteError(cmd, "character_read_error", "failed to read character: "+err.Error())
	}

	// 11. Resolve reference URI
	referenceURI, err := shared.ResolveMediaURI(flags.reference, "video")
	if err != nil {
		return common.WriteError(cmd, "reference_read_error", "failed to read reference: "+err.Error())
	}

	// 12. Build request body
	body := map[string]any{
		"model": "act_two",
		"character": map[string]any{
			"type": flags.characterType,
			"uri":  characterURI,
		},
		"reference": map[string]any{
			"type": "video",
			"uri":  referenceURI,
		},
		"bodyControl":         flags.bodyControl,
		"expressionIntensity": flags.expression,
		"ratio":               flags.ratio,
	}
	if flags.seed >= 0 {
		body["seed"] = flags.seed
	}
	if flags.publicFigure != "auto" {
		body["contentModeration"] = map[string]string{
			"publicFigureThreshold": flags.publicFigure,
		}
	}

	// 13. Make API request
	bodyJSON, _ := json.Marshal(body)
	req, err := shared.CreateRequest("POST", "/v1/character_performance", bytes.NewReader(bodyJSON))
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

	// 14. Parse response
	var taskResp shared.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return common.WriteError(cmd, "parse_error", "failed to parse response: "+err.Error())
	}

	// 15. Return task ID
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskResp.ID,
	})
}
