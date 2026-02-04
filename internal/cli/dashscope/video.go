package dashscope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

const (
	defaultBaseURL = "https://dashscope.aliyuncs.com/api/v1"

	// API paths
	videoSynthesisPath = "/services/aigc/video-generation/video-synthesis"
	kf2vSynthesisPath  = "/services/aigc/image2video/video-synthesis"
	taskQueryPath      = "/tasks/"
)

// Valid values
var (
	validResolutions = map[string]bool{
		"480P":  true,
		"720P":  true,
		"1080P": true,
	}
	validRatios = map[string]bool{
		"16:9": true,
		"9:16": true,
	}
	// All valid model names grouped by type
	validT2VModels = map[string]bool{
		"wan2.6-t2v":         true,
		"wan2.5-t2v-preview": true,
		"wan2.2-t2v-plus":    true,
		"wanx2.1-t2v-turbo":  true,
		"wanx2.1-t2v-plus":   true,
	}
	validI2VModels = map[string]bool{
		"wan2.6-i2v-flash":   true,
		"wan2.6-i2v":         true,
		"wan2.5-i2v-preview": true,
		"wan2.2-i2v-flash":   true,
		"wan2.2-i2v-plus":    true,
		"wanx2.1-i2v-turbo":  true,
		"wanx2.1-i2v-plus":   true,
	}
	validR2VModels = map[string]bool{
		"wan2.6-r2v-flash": true,
		"wan2.6-r2v":       true,
	}
	validKF2VModels = map[string]bool{
		"wan2.2-kf2v-flash":  true,
		"wanx2.1-kf2v-plus": true,
	}

	// Resolution to size mapping for t2v/r2v (which use "size" parameter)
	resolutionToSize = map[string]map[string]string{
		"16:9": {
			"480P":  "832*480",
			"720P":  "1280*720",
			"1080P": "1920*1080",
		},
		"9:16": {
			"480P":  "480*832",
			"720P":  "720*1280",
			"1080P": "1080*1920",
		},
	}

	// Models that support shot_type
	wan26Models = map[string]bool{
		"wan2.6-t2v":       true,
		"wan2.6-i2v-flash": true,
		"wan2.6-i2v":       true,
		"wan2.6-r2v-flash": true,
		"wan2.6-r2v":       true,
	}

	// Models that support audio_url
	audioURLModels = map[string]bool{
		"wan2.6-t2v":         true,
		"wan2.5-t2v-preview": true,
		"wan2.6-i2v-flash":   true,
		"wan2.6-i2v":         true,
		"wan2.5-i2v-preview": true,
	}
)

// Flag structs
type videoCreateFlags struct {
	image        string
	refs         []string
	firstFrame   string
	lastFrame    string
	promptFile   string
	model        string
	resolution   string
	ratio        string
	duration     int
	negative     string
	audio        bool
	noAudio      bool
	audioURL     string
	shotType     string
	promptExtend bool
	watermark    bool
	seed         int
}

// Commands
var videoCmd = newVideoCmd()

func newVideoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Generate videos using Tongyi Wanxiang",
		Long:  "Generate videos using Alibaba Tongyi Wanxiang models via DashScope API.",
	}

	cmd.AddCommand(newVideoCreateCmd())
	cmd.AddCommand(newVideoStatusCmd())
	cmd.AddCommand(newVideoDownloadCmd())

	return cmd
}

// ===== Create Command =====

func newVideoCreateCmd() *cobra.Command {
	flags := &videoCreateFlags{}

	cmd := &cobra.Command{
		Use:           "create [prompt]",
		Short:         "Create a video generation task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "Input image for i2v (local path or URL)")
	cmd.Flags().StringArrayVar(&flags.refs, "ref", nil, "Reference file(s) for r2v (repeatable, max 5)")
	cmd.Flags().StringVar(&flags.firstFrame, "first-frame", "", "First frame image for kf2v (local path or URL)")
	cmd.Flags().StringVar(&flags.lastFrame, "last-frame", "", "Last frame image for kf2v (requires --first-frame)")
	cmd.Flags().StringVarP(&flags.promptFile, "prompt-file", "f", "", "Read prompt from file")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "", "Model name (auto-selected based on input type)")
	cmd.Flags().StringVarP(&flags.resolution, "resolution", "r", "720P", "Resolution: 480P, 720P, 1080P")
	cmd.Flags().StringVar(&flags.ratio, "ratio", "16:9", "Aspect ratio: 16:9, 9:16 (t2v/r2v only)")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 5, "Duration in seconds")
	cmd.Flags().StringVar(&flags.negative, "negative", "", "Negative prompt (max 500 chars)")
	cmd.Flags().BoolVar(&flags.audio, "audio", false, "Enable auto audio generation (wan2.6-i2v-flash only)")
	cmd.Flags().BoolVar(&flags.noAudio, "no-audio", false, "Disable audio for r2v-flash (r2v has audio on by default)")
	cmd.Flags().StringVar(&flags.audioURL, "audio-url", "", "Custom audio URL (wav/mp3, 3-30s)")
	cmd.Flags().StringVar(&flags.shotType, "shot-type", "", "Shot type: single, multi (wan2.6 only)")
	cmd.Flags().BoolVar(&flags.promptExtend, "prompt-extend", true, "Enable prompt smart rewriting")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Add AI generated watermark")
	cmd.Flags().IntVar(&flags.seed, "seed", 0, "Random seed [0, 2147483647]")

	return cmd
}

// inputMode determines which API to use based on flags
type inputMode int

const (
	modeT2V inputMode = iota
	modeI2V
	modeR2V
	modeKF2V
)

func detectInputMode(flags *videoCreateFlags) inputMode {
	if flags.image != "" {
		return modeI2V
	}
	if len(flags.refs) > 0 {
		return modeR2V
	}
	if flags.firstFrame != "" {
		return modeKF2V
	}
	return modeT2V
}

func defaultModel(mode inputMode) string {
	switch mode {
	case modeI2V:
		return "wan2.6-i2v-flash"
	case modeR2V:
		return "wan2.6-r2v-flash"
	case modeKF2V:
		return "wan2.2-kf2v-flash"
	default:
		return "wan2.6-t2v"
	}
}

func isValidModel(model string, mode inputMode) bool {
	switch mode {
	case modeI2V:
		return validI2VModels[model]
	case modeR2V:
		return validR2VModels[model]
	case modeKF2V:
		return validKF2VModels[model]
	default:
		return validT2VModels[model]
	}
}

func runVideoCreate(cmd *cobra.Command, args []string, flags *videoCreateFlags) error {
	// Get prompt
	prompt, err := getVideoPrompt(args, flags.promptFile, cmd.InOrStdin())
	if err != nil {
		return common.WriteError(cmd, "missing_prompt", err.Error())
	}

	// Check conflicting input flags
	inputCount := 0
	if flags.image != "" {
		inputCount++
	}
	if len(flags.refs) > 0 {
		inputCount++
	}
	if flags.firstFrame != "" {
		inputCount++
	}
	if inputCount > 1 {
		return common.WriteError(cmd, "conflicting_input_flags", "cannot use --image, --ref, and --first-frame together")
	}

	// Detect input mode and resolve model
	mode := detectInputMode(flags)
	model := flags.model
	if model == "" {
		model = defaultModel(mode)
	}

	// Validate model
	if !isValidModel(model, mode) {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s' for this input type", model))
	}

	// Validate resolution
	if !validResolutions[flags.resolution] {
		return common.WriteError(cmd, "invalid_resolution", fmt.Sprintf("invalid resolution '%s', use 480P, 720P, or 1080P", flags.resolution))
	}

	// Validate ratio (only for t2v and r2v)
	if mode == modeT2V || mode == modeR2V {
		if !validRatios[flags.ratio] {
			return common.WriteError(cmd, "invalid_ratio", fmt.Sprintf("invalid ratio '%s', use 16:9 or 9:16", flags.ratio))
		}
	}

	// Validate duration
	if flags.duration < 2 || flags.duration > 15 {
		return common.WriteError(cmd, "invalid_duration", "duration must be between 2 and 15 seconds")
	}

	// Validate image input
	if flags.image != "" && !isURL(flags.image) {
		if _, err := os.Stat(flags.image); os.IsNotExist(err) {
			return common.WriteError(cmd, "image_not_found", fmt.Sprintf("image not found: %s", flags.image))
		}
	}

	// Validate first frame
	if flags.firstFrame != "" && !isURL(flags.firstFrame) {
		if _, err := os.Stat(flags.firstFrame); os.IsNotExist(err) {
			return common.WriteError(cmd, "first_frame_not_found", fmt.Sprintf("first frame not found: %s", flags.firstFrame))
		}
	}

	// Validate last frame
	if flags.lastFrame != "" {
		if flags.firstFrame == "" {
			return common.WriteError(cmd, "last_frame_requires_first", "--last-frame requires --first-frame")
		}
		if !isURL(flags.lastFrame) {
			if _, err := os.Stat(flags.lastFrame); os.IsNotExist(err) {
				return common.WriteError(cmd, "last_frame_not_found", fmt.Sprintf("last frame not found: %s", flags.lastFrame))
			}
		}
	}

	// Validate refs
	if len(flags.refs) > 5 {
		return common.WriteError(cmd, "too_many_refs", "too many reference files (max 5)")
	}
	for _, ref := range flags.refs {
		if !isURL(ref) {
			if _, err := os.Stat(ref); os.IsNotExist(err) {
				return common.WriteError(cmd, "ref_not_found", fmt.Sprintf("reference file not found: %s", ref))
			}
		}
	}

	// Validate compatibility: --audio only with wan2.6-i2v-flash
	if flags.audio && model != "wan2.6-i2v-flash" {
		return common.WriteError(cmd, "incompatible_audio", "--audio only supported by wan2.6-i2v-flash")
	}

	// Validate compatibility: --no-audio only with wan2.6-r2v-flash
	if flags.noAudio && model != "wan2.6-r2v-flash" {
		return common.WriteError(cmd, "incompatible_no_audio", "--no-audio only supported by wan2.6-r2v-flash")
	}

	// Validate compatibility: --shot-type only with wan2.6 models
	if flags.shotType != "" && !wan26Models[model] {
		return common.WriteError(cmd, "incompatible_shot_type", "--shot-type only supported by wan2.6 models")
	}

	// Validate compatibility: --audio-url not for r2v models
	if flags.audioURL != "" {
		if validR2VModels[model] {
			return common.WriteError(cmd, "incompatible_audio_url", "--audio-url not supported by r2v models")
		}
		if !audioURLModels[model] {
			return common.WriteError(cmd, "incompatible_audio_url", fmt.Sprintf("--audio-url not supported by model '%s'", model))
		}
	}

	// Check API key
	apiKey := config.GetAPIKey("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("DASHSCOPE_API_KEY"))
	}

	// Build request body
	input := map[string]any{
		"prompt": prompt,
	}

	if flags.negative != "" {
		input["negative_prompt"] = flags.negative
	}

	// Mode-specific input fields
	switch mode {
	case modeI2V:
		input["img_url"] = flags.image
		if flags.audioURL != "" {
			input["audio_url"] = flags.audioURL
		}
	case modeR2V:
		input["reference_urls"] = flags.refs
	case modeKF2V:
		input["first_frame_url"] = flags.firstFrame
		if flags.lastFrame != "" {
			input["last_frame_url"] = flags.lastFrame
		}
	default: // t2v
		if flags.audioURL != "" {
			input["audio_url"] = flags.audioURL
		}
	}

	params := map[string]any{
		"prompt_extend": flags.promptExtend,
		"watermark":     flags.watermark,
	}

	// Resolution/size parameter
	switch mode {
	case modeT2V, modeR2V:
		// Use "size" parameter with pixel values
		sizeMap := resolutionToSize[flags.ratio]
		if sizeMap != nil {
			params["size"] = sizeMap[flags.resolution]
		}
		params["duration"] = flags.duration
	case modeI2V:
		params["resolution"] = flags.resolution
		params["duration"] = flags.duration
	case modeKF2V:
		params["resolution"] = flags.resolution
		// kf2v duration is fixed at 5
	}

	// Optional parameters
	if flags.shotType != "" {
		params["shot_type"] = flags.shotType
	}
	if flags.seed != 0 {
		params["seed"] = flags.seed
	}
	if mode == modeI2V && flags.audio {
		params["audio"] = true
	}
	if mode == modeR2V && flags.noAudio {
		params["audio"] = false
	}

	body := map[string]any{
		"model":      model,
		"input":      input,
		"parameters": params,
	}

	// Determine API path
	apiPath := videoSynthesisPath
	if mode == modeKF2V {
		apiPath = kf2vSynthesisPath
	}

	// Send request
	baseURL := getBaseURL()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := http.NewRequest("POST", baseURL+apiPath, bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-DashScope-Async", "enable")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	var result struct {
		Output *struct {
			TaskID     string `json:"task_id"`
			TaskStatus string `json:"task_status"`
		} `json:"output"`
		RequestID string `json:"request_id"`
		Code      string `json:"code"`
		Message   string `json:"message"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != "" {
		return common.WriteError(cmd, result.Code, result.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return handleHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	taskID := ""
	status := "pending"
	if result.Output != nil {
		taskID = result.Output.TaskID
		status = strings.ToLower(result.Output.TaskStatus)
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"status":  status,
	})
}

// ===== Status Command =====

func newVideoStatusCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Get video generation status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoStatus(cmd, args, verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full output including video URL")

	return cmd
}

func runVideoStatus(cmd *cobra.Command, args []string, verbose bool) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	apiKey := config.GetAPIKey("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("DASHSCOPE_API_KEY"))
	}

	baseURL := getBaseURL()
	req, err := http.NewRequest("GET", baseURL+taskQueryPath+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	var result struct {
		Output *struct {
			TaskID       string `json:"task_id"`
			TaskStatus   string `json:"task_status"`
			VideoURL     string `json:"video_url"`
			OrigPrompt   string `json:"orig_prompt"`
			ActualPrompt string `json:"actual_prompt"`
			Message      string `json:"message"`
		} `json:"output"`
		Usage *struct {
			Duration            int `json:"duration"`
			OutputVideoDuration int `json:"output_video_duration"`
			SR                  int `json:"SR"`
		} `json:"usage"`
		RequestID string `json:"request_id"`
		Code      string `json:"code"`
		Message   string `json:"message"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != "" {
		return common.WriteError(cmd, result.Code, result.Message)
	}

	if result.Output == nil {
		return common.WriteError(cmd, "response_error", "empty response from API")
	}

	status := strings.ToLower(result.Output.TaskStatus)

	if status == "failed" {
		msg := result.Output.Message
		if msg == "" {
			msg = "video generation failed"
		}
		return common.WriteError(cmd, "video_failed", msg)
	}

	output := map[string]any{
		"success": true,
		"task_id": result.Output.TaskID,
		"status":  status,
	}

	if status == "succeeded" {
		if result.Usage != nil {
			output["duration"] = result.Usage.OutputVideoDuration
			output["resolution"] = result.Usage.SR
		}
		if verbose {
			output["video_url"] = result.Output.VideoURL
			if result.Output.OrigPrompt != "" {
				output["orig_prompt"] = result.Output.OrigPrompt
			}
			if result.Output.ActualPrompt != "" {
				output["actual_prompt"] = result.Output.ActualPrompt
			}
		}
	}

	return common.WriteSuccess(cmd, output)
}

// ===== Download Command =====

func newVideoDownloadCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "download <task_id>",
		Short:         "Download completed video",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoDownload(cmd, args, output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (.mp4)")

	return cmd
}

func runVideoDownload(cmd *cobra.Command, args []string, output string) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	if output == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}

	ext := strings.ToLower(filepath.Ext(output))
	if ext != ".mp4" {
		return common.WriteError(cmd, "invalid_format", fmt.Sprintf("unsupported format '%s', use .mp4", ext))
	}

	apiKey := config.GetAPIKey("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("DASHSCOPE_API_KEY"))
	}

	// Query task status first
	baseURL := getBaseURL()
	req, err := http.NewRequest("GET", baseURL+taskQueryPath+taskID, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	var result struct {
		Output *struct {
			TaskStatus string `json:"task_status"`
			VideoURL   string `json:"video_url"`
		} `json:"output"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != "" {
		return common.WriteError(cmd, result.Code, result.Message)
	}

	if result.Output == nil {
		return common.WriteError(cmd, "response_error", "empty response from API")
	}

	status := strings.ToLower(result.Output.TaskStatus)
	if status == "failed" {
		return common.WriteError(cmd, "video_failed", "video generation failed")
	}
	if status != "succeeded" {
		return common.WriteError(cmd, "video_not_ready", fmt.Sprintf("video is not ready, current status: %s", status))
	}

	videoURL := result.Output.VideoURL
	if videoURL == "" {
		return common.WriteError(cmd, "no_video", "no video URL in response")
	}

	// Download video
	dlResp, err := http.Get(videoURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download video: %s", err.Error()))
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		if dlResp.StatusCode == http.StatusForbidden || dlResp.StatusCode == http.StatusNotFound {
			return common.WriteError(cmd, "url_expired", "video URL has expired (24h limit)")
		}
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status %d", dlResp.StatusCode))
	}

	absOutput, err := filepath.Abs(output)
	if err != nil {
		absOutput = output
	}

	file, err := os.Create(absOutput)
	if err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot create output file: %s", err.Error()))
	}
	defer file.Close()

	if _, err := io.Copy(file, dlResp.Body); err != nil {
		return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write output file: %s", err.Error()))
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"file":    absOutput,
	})
}

// ===== Helper Functions =====

func getVideoPrompt(args []string, promptFile string, stdin io.Reader) (string, error) {
	// From positional argument
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0], nil
	}

	// From file
	if promptFile != "" {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			return "", fmt.Errorf("cannot read prompt file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", fmt.Errorf("prompt file is empty")
		}
		return text, nil
	}

	// From stdin
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("prompt is required")
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("prompt is required")
	}
	return text, nil
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func getBaseURL() string {
	if url := config.GetAPIKey("DASHSCOPE_BASE_URL"); url != "" {
		return url
	}
	return defaultBaseURL
}

func handleAPIError(cmd *cobra.Command, err error) error {
	if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
		return common.WriteError(cmd, "connection_error", fmt.Sprintf("cannot connect to DashScope API: %s", err.Error()))
	}
	if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
		return common.WriteError(cmd, "timeout", fmt.Sprintf("request timed out: %s", err.Error()))
	}
	return common.WriteError(cmd, "connection_error", err.Error())
}

func handleHTTPError(cmd *cobra.Command, statusCode int, body string) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid or region mismatch")
	case http.StatusTooManyRequests:
		return common.WriteError(cmd, "rate_limit", "too many requests")
	case http.StatusBadRequest:
		return common.WriteError(cmd, "invalid_request", fmt.Sprintf("invalid request: %s", body))
	default:
		return common.WriteError(cmd, "server_error", fmt.Sprintf("server error (HTTP %d): %s", statusCode, body))
	}
}
