package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type motionControlFlags struct {
	image       string
	video       string
	orientation string
	mode        string
	keepSound   bool
	watermark   bool
	promptFile  string
}

func newMotionControlCmd() *cobra.Command {
	flags := &motionControlFlags{}

	cmd := &cobra.Command{
		Use:           "create-motion-control [prompt]",
		Short:         "Create video with motion control (transfer motion from video to image)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMotionControl(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Reference image (required)")
	cmd.Flags().StringVarP(&flags.video, "video", "v", "", "Reference video for motion (required)")
	cmd.Flags().StringVarP(&flags.orientation, "orientation", "o", "image", "Character orientation: image, video")
	cmd.Flags().StringVarP(&flags.mode, "mode", "m", "std", "Generation mode: std, pro")
	cmd.Flags().BoolVar(&flags.keepSound, "keep-sound", true, "Keep original video sound")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Include watermark")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")

	return cmd
}

var validOrientations = map[string]bool{
	"image": true,
	"video": true,
}

func runMotionControl(cmd *cobra.Command, args []string, flags *motionControlFlags) error {
	// Validate image is required
	if flags.image == "" {
		return common.WriteError(cmd, "missing_image", "reference image is required (-i)")
	}

	// Validate video is required
	if flags.video == "" {
		return common.WriteError(cmd, "missing_video", "reference video is required (-v)")
	}

	// Validate image file exists (if local)
	if !isURL(flags.image) {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image not found: %s", flags.image))
		}
	}

	// Get prompt (optional)
	prompt, _ := getPrompt(args, flags.promptFile, cmd.InOrStdin())

	// Validate orientation
	if !validOrientations[flags.orientation] {
		return common.WriteError(cmd, "invalid_orientation", fmt.Sprintf("invalid orientation '%s', use image or video", flags.orientation))
	}

	// Validate mode
	if !validModes[flags.mode] {
		return common.WriteError(cmd, "invalid_mode", fmt.Sprintf("invalid mode '%s', use std or pro", flags.mode))
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := generateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Resolve image URL
	imageURL, err := resolveImageURL(flags.image)
	if err != nil {
		return common.WriteError(cmd, "image_read_error", fmt.Sprintf("cannot read image: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"image_url":             imageURL,
		"video_url":             flags.video,
		"character_orientation": flags.orientation,
		"mode":                  flags.mode,
	}

	if prompt != "" {
		body["prompt"] = prompt
	}

	if !flags.keepSound {
		body["keep_original_sound"] = "no"
	}

	if flags.watermark {
		body["watermark_info"] = map[string]bool{"enabled": true}
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/videos/motion-control", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			TaskID     string `json:"task_id"`
			TaskStatus string `json:"task_status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	})
}
