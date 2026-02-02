package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type createFlags struct {
	model           string
	promptFile      string
	duration        int
	resolution      string
	promptOptimizer bool
	fastPretreat    bool
	callbackURL     string
	firstFrame      string
	lastFrame       string
	subject         string
}

var validT2VModels = map[string]bool{
	"MiniMax-Hailuo-2.3": true,
	"MiniMax-Hailuo-02":  true,
	"T2V-01-Director":    true,
	"T2V-01":             true,
}

var validI2VModels = map[string]bool{
	"MiniMax-Hailuo-2.3":      true,
	"MiniMax-Hailuo-2.3-Fast": true,
	"MiniMax-Hailuo-02":       true,
	"I2V-01-Director":         true,
	"I2V-01-live":             true,
	"I2V-01":                  true,
}

// fl2v only supports MiniMax-Hailuo-02
const fl2vModel = "MiniMax-Hailuo-02"

// s2v only supports S2V-01
const s2vModel = "S2V-01"

var validResolutionsT2V = map[string]bool{
	"720P":  true,
	"768P":  true,
	"1080P": true,
}

var validResolutionsI2V = map[string]bool{
	"512P":  true,
	"720P":  true,
	"768P":  true,
	"1080P": true,
}

var validResolutionsFL2V = map[string]bool{
	"768P":  true,
	"1080P": true,
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create a video generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.model, "model", "m", "", "Model name (auto-selected if not specified)")
	cmd.Flags().StringVar(&flags.promptFile, "prompt-file", "", "Read prompt from file")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 6, "Video duration in seconds (typically 6 or 10)")
	cmd.Flags().StringVarP(&flags.resolution, "resolution", "r", "", "Resolution: 720P/768P/1080P (type-specific)")
	cmd.Flags().BoolVar(&flags.promptOptimizer, "prompt-optimizer", true, "Enable prompt optimization")
	cmd.Flags().BoolVar(&flags.fastPretreat, "fast-pretreatment", false, "Enable fast pretreatment (Hailuo models only)")
	cmd.Flags().StringVar(&flags.callbackURL, "callback-url", "", "Callback URL for task status updates")
	cmd.Flags().StringVar(&flags.firstFrame, "first-frame", "", "First frame image (URL or local file)")
	cmd.Flags().StringVar(&flags.lastFrame, "last-frame", "", "Last frame image (URL or local file)")
	cmd.Flags().StringVar(&flags.subject, "subject", "", "Subject reference image (URL or local file)")

	return cmd
}

type createResponse struct {
	Success bool   `json:"success"`
	TaskID  string `json:"task_id,omitempty"`
	Model   string `json:"model,omitempty"`
	Type    string `json:"type,omitempty"`
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	prompt, err := getPromptOptional(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "prompt_read_error", err.Error())
	}

	// Auto-detect generation type based on parameters
	genType := detectGenType(flags)

	// Validate and set model
	model := flags.model
	switch genType {
	case "s2v":
		// s2v only supports S2V-01
		model = s2vModel
		if flags.resolution != "" {
			return common.WriteError(cmd, "invalid_parameter", "resolution is not supported for subject reference mode")
		}
		if flags.duration != 6 {
			return common.WriteError(cmd, "invalid_parameter", "duration is not supported for subject reference mode")
		}
	case "fl2v":
		// fl2v only supports MiniMax-Hailuo-02
		model = fl2vModel
		if flags.resolution != "" && !validResolutionsFL2V[flags.resolution] {
			return common.WriteError(cmd, "invalid_resolution", "resolution must be 768P or 1080P for first-last frame mode")
		}
	case "i2v":
		if model == "" {
			model = "MiniMax-Hailuo-2.3"
		}
		if !validI2VModels[model] {
			return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s' for image-to-video", model))
		}
		if flags.resolution != "" && !validResolutionsI2V[flags.resolution] {
			return common.WriteError(cmd, "invalid_resolution", "resolution must be 512P, 720P, 768P, or 1080P for image-to-video")
		}
	case "t2v":
		if strings.TrimSpace(prompt) == "" {
			return common.WriteError(cmd, "missing_prompt", "prompt is required for text-to-video")
		}
		if model == "" {
			model = "MiniMax-Hailuo-2.3"
		}
		if !validT2VModels[model] {
			return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s' for text-to-video", model))
		}
		if flags.resolution != "" && !validResolutionsT2V[flags.resolution] {
			return common.WriteError(cmd, "invalid_resolution", "resolution must be 720P, 768P, or 1080P for text-to-video")
		}
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := map[string]any{
		"model":            model,
		"prompt_optimizer": flags.promptOptimizer,
	}
	if strings.TrimSpace(prompt) != "" {
		body["prompt"] = prompt
	}
	if flags.fastPretreat {
		body["fast_pretreatment"] = true
	}
	if flags.callbackURL != "" {
		body["callback_url"] = flags.callbackURL
	}
	if flags.duration > 0 && genType != "s2v" {
		body["duration"] = flags.duration
	}
	if flags.resolution != "" && genType != "s2v" {
		body["resolution"] = flags.resolution
	}

	switch genType {
	case "i2v":
		first, err := shared.ResolveImageURL(flags.firstFrame)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read first-frame: %s", err.Error()))
		}
		body["first_frame_image"] = first
	case "fl2v":
		first, err := shared.ResolveImageURL(flags.firstFrame)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read first-frame: %s", err.Error()))
		}
		last, err := shared.ResolveImageURL(flags.lastFrame)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read last-frame: %s", err.Error()))
		}
		body["first_frame_image"] = first
		body["last_frame_image"] = last
	case "s2v":
		subject, err := shared.ResolveImageURL(flags.subject)
		if err != nil {
			return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read subject image: %s", err.Error()))
		}
		body["subject_reference"] = []map[string]any{
			{
				"type":  "character",
				"image": []string{subject},
			},
		}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := shared.CreateRequest("POST", "/v1/video_generation", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}

	resp, err := shared.DoRequest(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var apiResp struct {
		TaskID   string `json:"task_id"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg))
	}

	return common.WriteSuccess(cmd, createResponse{
		Success: true,
		TaskID:  apiResp.TaskID,
		Model:   model,
		Type:    genType,
	})
}

// detectGenType infers generation type from flags
func detectGenType(flags *createFlags) string {
	if flags.subject != "" {
		return "s2v"
	}
	if flags.firstFrame != "" && flags.lastFrame != "" {
		return "fl2v"
	}
	if flags.firstFrame != "" {
		return "i2v"
	}
	return "t2v"
}

func getPromptOptional(args []string, filePath string, stdin io.Reader) (string, error) {
	if len(args) > 0 {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text != "" {
			return text, nil
		}
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		return text, nil
	}

	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", nil
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		return text, nil
	}

	return "", nil
}
