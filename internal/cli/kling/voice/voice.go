package voice

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/kling/video"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var Cmd = NewCmd()

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voice",
		Short: "Manage custom voices for video generation",
		Long:  "Create, list, and delete custom voices for use in video generation.",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}

// =============================================================================
// Create Command
// =============================================================================

type createFlags struct {
	audio   string
	videoID string
}

func newCreateCmd() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:           "create <name>",
		Short:         "Create a custom voice from audio or video",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.audio, "audio", "a", "", "Audio file for voice cloning (local file or URL, 5-30s)")
	cmd.Flags().StringVar(&flags.videoID, "video-id", "", "Video ID from v2.6/avatar/lip-sync generation")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string, flags *createFlags) error {
	// Validate name
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_name", "voice name is required")
	}
	name := args[0]

	if len(name) > 20 {
		return common.WriteError(cmd, "invalid_name", "voice name must be at most 20 characters")
	}

	// Validate audio source (audio or video-id required, but not both)
	if flags.audio == "" && flags.videoID == "" {
		return common.WriteError(cmd, "missing_audio", "audio file (--audio) or video ID (--video-id) is required")
	}
	if flags.audio != "" && flags.videoID != "" {
		return common.WriteError(cmd, "conflicting_source", "cannot use both --audio and --video-id")
	}

	// Validate audio file exists (if local)
	if flags.audio != "" && !isURL(flags.audio) {
		if _, err := os.Stat(flags.audio); os.IsNotExist(err) {
			return common.WriteError(cmd, "audio_not_found", fmt.Sprintf("audio file not found: %s", flags.audio))
		}
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"voice_name": name,
	}

	if flags.videoID != "" {
		body["video_id"] = flags.videoID
	} else {
		// Resolve audio URL
		audioURL, err := resolveAudioURL(flags.audio)
		if err != nil {
			return common.WriteError(cmd, "audio_read_error", fmt.Sprintf("cannot read audio: %s", err.Error()))
		}
		body["voice_url"] = audioURL
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", video.GetKlingAPIBase()+"/v1/general/custom-voices", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return video.HandleAPIError(cmd, err)
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
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success":    true,
		"task_id":    result.Data.TaskID,
		"status":     result.Data.TaskStatus,
		"voice_name": name,
	})
}

// =============================================================================
// Status Command
// =============================================================================

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Get voice creation status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args)
		},
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Validate task ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", video.GetKlingAPIBase()+"/v1/general/custom-voices/"+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return video.HandleAPIError(cmd, err)
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
			TaskID        string `json:"task_id"`
			TaskStatus    string `json:"task_status"`
			TaskStatusMsg string `json:"task_status_msg"`
			TaskResult    *struct {
				Voices []struct {
					VoiceID   string `json:"voice_id"`
					VoiceName string `json:"voice_name"`
					TrialURL  string `json:"trial_url"`
					OwnedBy   string `json:"owned_by"`
				} `json:"voices"`
			} `json:"task_result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	// Handle failed status
	if result.Data.TaskStatus == "failed" {
		msg := result.Data.TaskStatusMsg
		if msg == "" {
			msg = "voice creation failed"
		}
		return common.WriteError(cmd, "voice_failed", msg)
	}

	// Build response
	output := map[string]any{
		"success": true,
		"task_id": result.Data.TaskID,
		"status":  result.Data.TaskStatus,
	}

	if result.Data.TaskStatus == "succeed" && result.Data.TaskResult != nil && len(result.Data.TaskResult.Voices) > 0 {
		voice := result.Data.TaskResult.Voices[0]
		output["voice_id"] = voice.VoiceID
		output["voice_name"] = voice.VoiceName
		output["trial_url"] = voice.TrialURL
	}

	return common.WriteSuccess(cmd, output)
}

// =============================================================================
// List Command
// =============================================================================

type listFlags struct {
	voiceType string
	limit     int
	page      int
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List voices",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.voiceType, "type", "t", "custom", "Voice type: custom, official, tts")
	cmd.Flags().IntVarP(&flags.limit, "limit", "l", 30, "Maximum voices to return (1-500)")
	cmd.Flags().IntVarP(&flags.page, "page", "p", 1, "Page number")

	return cmd
}

// TTS preset voices (bilingual: zh/en)
var ttsVoices = []map[string]any{
	{"voice_id": "genshin_vindi2", "voice_name": "阳光少年 / Sunny", "category": "Young Male", "languages": []string{"zh", "en"}},
	{"voice_id": "zhinen_xuesheng", "voice_name": "懂事小弟 / Sage", "category": "Young Male", "languages": []string{"zh", "en"}},
	{"voice_id": "ai_kaiya", "voice_name": "阳光男生 / Shine", "category": "Young Male", "languages": []string{"zh", "en"}},
	{"voice_id": "ai_chenjiahao_712", "voice_name": "文艺小哥 / Lyric", "category": "Young Male", "languages": []string{"zh", "en"}},
	{"voice_id": "ai_shatang", "voice_name": "青春少女 / Blossom", "category": "Young Female", "languages": []string{"zh", "en"}},
	{"voice_id": "genshin_klee2", "voice_name": "温柔小妹 / Peppy", "category": "Young Female", "languages": []string{"zh", "en"}},
	{"voice_id": "genshin_kirara", "voice_name": "元气少女 / Dove", "category": "Young Female", "languages": []string{"zh", "en"}},
	{"voice_id": "chat1_female_new-3", "voice_name": "温柔姐姐 / Tender", "category": "Adult Female", "languages": []string{"zh", "en"}},
	{"voice_id": "chengshu_jiejie", "voice_name": "优雅贵妇 / Grace", "category": "Adult Female", "languages": []string{"zh", "en"}},
	{"voice_id": "you_pingjing", "voice_name": "温柔妈妈 / Helen", "category": "Adult Female", "languages": []string{"zh", "en"}},
	{"voice_id": "ai_huangyaoshi_712", "voice_name": "稳重老爸 / Rock", "category": "Adult Male", "languages": []string{"zh", "en"}},
	{"voice_id": "ai_laoguowang_712", "voice_name": "严肃上司 / Titan", "category": "Adult Male", "languages": []string{"zh", "en"}},
	{"voice_id": "cartoon-boy-07", "voice_name": "活泼男童 / Zippy", "category": "Child", "languages": []string{"zh", "en"}},
	{"voice_id": "cartoon-girl-01", "voice_name": "俏皮女童 / Sprite", "category": "Child", "languages": []string{"zh", "en"}},
	{"voice_id": "laopopo_speech02", "voice_name": "唠叨奶奶 / Prattle", "category": "Elderly", "languages": []string{"zh", "en"}},
	{"voice_id": "heainainai_speech02", "voice_name": "和蔼奶奶 / Hearth", "category": "Elderly", "languages": []string{"zh", "en"}},
}

func runList(cmd *cobra.Command, flags *listFlags) error {
	// Validate type
	if flags.voiceType != "custom" && flags.voiceType != "official" && flags.voiceType != "tts" {
		return common.WriteError(cmd, "invalid_type", "type must be 'custom', 'official', or 'tts'")
	}

	// Handle TTS voices (no API call needed)
	if flags.voiceType == "tts" {
		return common.WriteSuccess(cmd, map[string]any{
			"success": true,
			"type":    "tts",
			"voices":  ttsVoices,
			"count":   len(ttsVoices),
		})
	}

	// Validate limit
	if flags.limit < 1 || flags.limit > 500 {
		return common.WriteError(cmd, "invalid_limit", "limit must be between 1 and 500")
	}

	// Validate page
	if flags.page < 1 {
		return common.WriteError(cmd, "invalid_page", "page must be at least 1")
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Determine endpoint
	endpoint := "/v1/general/custom-voices"
	if flags.voiceType == "official" {
		endpoint = "/v1/general/presets-voices"
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s?pageNum=%d&pageSize=%d", video.GetKlingAPIBase(), endpoint, flags.page, flags.limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return video.HandleAPIError(cmd, err)
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
		Data    []struct {
			TaskID     string `json:"task_id"`
			TaskStatus string `json:"task_status"`
			TaskResult *struct {
				Voices []struct {
					VoiceID   string `json:"voice_id"`
					VoiceName string `json:"voice_name"`
					TrialURL  string `json:"trial_url"`
					OwnedBy   string `json:"owned_by"`
				} `json:"voices"`
			} `json:"task_result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	// Build response - flatten voices from tasks
	voices := []map[string]any{}
	if result.Data != nil {
		for _, task := range result.Data {
			if task.TaskResult != nil && task.TaskStatus == "succeed" {
				for _, v := range task.TaskResult.Voices {
					voices = append(voices, map[string]any{
						"voice_id":   v.VoiceID,
						"voice_name": v.VoiceName,
						"trial_url":  v.TrialURL,
						"owned_by":   v.OwnedBy,
					})
				}
			}
		}
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"type":    flags.voiceType,
		"voices":  voices,
		"count":   len(voices),
	})
}

// =============================================================================
// Delete Command
// =============================================================================

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete <voice_id>",
		Short:         "Delete a custom voice",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args)
		},
	}

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	// Validate voice ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_voice_id", "voice ID is required")
	}
	voiceID := args[0]

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"voice_id": voiceID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", video.GetKlingAPIBase()+"/v1/general/delete-voices", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return video.HandleAPIError(cmd, err)
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
		return video.HandleKlingError(cmd, result.Code, result.Message)
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success":  true,
		"voice_id": voiceID,
		"status":   "deleted",
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// isURL checks if the input looks like a URL.
func isURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

// resolveAudioURL returns the audio URL for API request.
// If input is URL, returns as-is. If local file, encodes to base64.
func resolveAudioURL(input string) (string, error) {
	if isURL(input) {
		return input, nil
	}
	return encodeAudioToBase64(input)
}

// encodeAudioToBase64 encodes local audio file to pure base64 string.
func encodeAudioToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
