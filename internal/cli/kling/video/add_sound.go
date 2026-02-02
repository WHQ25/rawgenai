package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type addSoundFlags struct {
	url   string
	sound string
	bgm   string
	asmr  bool
}

func newAddSoundCmd() *cobra.Command {
	flags := &addSoundFlags{}

	cmd := &cobra.Command{
		Use:           "add-sound [video_id]",
		Short:         "Add sound effects or BGM to video",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddSound(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.url, "url", "", "Video URL (alternative to video_id)")
	cmd.Flags().StringVarP(&flags.sound, "sound", "s", "", "Sound effect prompt (max 200 chars)")
	cmd.Flags().StringVarP(&flags.bgm, "bgm", "b", "", "Background music prompt (max 200 chars)")
	cmd.Flags().BoolVar(&flags.asmr, "asmr", false, "Enable ASMR mode (enhanced detail)")

	return cmd
}

func runAddSound(cmd *cobra.Command, args []string, flags *addSoundFlags) error {
	// Validate: either video_id or url must be provided
	videoID := ""
	if len(args) > 0 {
		videoID = strings.TrimSpace(args[0])
	}

	if videoID == "" && flags.url == "" {
		return common.WriteError(cmd, "missing_video_id", "video ID or --url is required")
	}

	if videoID != "" && flags.url != "" {
		return common.WriteError(cmd, "conflicting_input", "cannot use both video_id and --url")
	}

	// Validate sound prompt length
	if len(flags.sound) > 200 {
		return common.WriteError(cmd, "invalid_sound", "sound prompt must be at most 200 characters")
	}

	// Validate bgm prompt length
	if len(flags.bgm) > 200 {
		return common.WriteError(cmd, "invalid_bgm", "bgm prompt must be at most 200 characters")
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

	// Build request body
	body := map[string]any{}

	if videoID != "" {
		body["video_id"] = videoID
	} else {
		body["video_url"] = flags.url
	}

	// Sound effect configuration
	if flags.sound != "" || flags.asmr {
		foleyConfig := map[string]any{}
		if flags.sound != "" {
			foleyConfig["prompt"] = flags.sound
		}
		if flags.asmr {
			foleyConfig["asmr_detail_mode"] = "open"
		}
		body["foley_config"] = foleyConfig
	}

	// BGM configuration
	if flags.bgm != "" {
		body["bgm_config"] = map[string]any{
			"prompt": flags.bgm,
		}
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/audio/video-to-audio", bytes.NewReader(jsonBody))
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

	// Check for errors
	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	// Return success
	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	})
}
