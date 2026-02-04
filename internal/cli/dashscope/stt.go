package dashscope

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

const (
	sttTranscriptionPath = "/services/audio/asr/transcription"
	sttMaxSyncFileSize   = 10 * 1024 * 1024 // 10 MB
)

// Valid STT models
var (
	validSyncSTTModels = map[string]bool{
		"qwen3-asr-flash":            true,
		"qwen3-asr-flash-2025-09-08": true,
	}

	validRunTaskSTTModels = map[string]bool{
		"paraformer-realtime-v2":      true,
		"paraformer-realtime-v1":      true,
		"paraformer-realtime-8k-v2":   true,
		"paraformer-realtime-8k-v1":   true,
		"fun-asr-realtime":            true,
		"fun-asr-realtime-2025-11-07": true,
		"fun-asr-realtime-2025-09-15": true,
	}

	validSessionUpdateSTTModels = map[string]bool{
		"qwen3-asr-flash-realtime":            true,
		"qwen3-asr-flash-realtime-2025-10-27": true,
	}

	validAsyncSTTModels = map[string]bool{
		"paraformer-v2":              true,
		"paraformer-v1":              true,
		"paraformer-8k-v2":          true,
		"paraformer-8k-v1":          true,
		"paraformer-mtl-v1":         true,
		"fun-asr":                   true,
		"fun-asr-mtl":               true,
		"qwen3-asr-flash-filetrans": true,
	}

	// Models that support --vocabulary-id (realtime)
	sttVocabularyModels = map[string]bool{
		"paraformer-realtime-v2":      true,
		"paraformer-realtime-v1":      true,
		"paraformer-realtime-8k-v2":   true,
		"paraformer-realtime-8k-v1":   true,
		"fun-asr-realtime":            true,
		"fun-asr-realtime-2025-11-07": true,
		"fun-asr-realtime-2025-09-15": true,
	}

	// Models that support --disfluency-removal (realtime)
	sttDisfluencyModels = map[string]bool{
		"paraformer-realtime-v2":      true,
		"paraformer-realtime-v1":      true,
		"paraformer-realtime-8k-v2":   true,
		"paraformer-realtime-8k-v1":   true,
		"fun-asr-realtime":            true,
		"fun-asr-realtime-2025-11-07": true,
		"fun-asr-realtime-2025-09-15": true,
	}

	// Models that support --language-hints (realtime)
	sttLanguageHintsRealtimeModels = map[string]bool{
		"paraformer-realtime-v2":      true,
		"fun-asr-realtime":            true,
		"fun-asr-realtime-2025-11-07": true,
		"fun-asr-realtime-2025-09-15": true,
	}

	// Async models that support --vocabulary-id
	sttAsyncVocabularyModels = map[string]bool{
		"paraformer-v2":      true,
		"paraformer-v1":      true,
		"paraformer-8k-v2":  true,
		"paraformer-8k-v1":  true,
		"paraformer-mtl-v1": true,
		"fun-asr":           true,
		"fun-asr-mtl":       true,
	}

	// Async models that support --disfluency-removal
	sttAsyncDisfluencyModels = map[string]bool{
		"paraformer-v2":      true,
		"paraformer-v1":      true,
		"paraformer-8k-v2":  true,
		"paraformer-8k-v1":  true,
		"paraformer-mtl-v1": true,
		"fun-asr":           true,
		"fun-asr-mtl":       true,
	}

	// Async models that support --diarize
	sttAsyncDiarizeModels = map[string]bool{
		"paraformer-v2":      true,
		"paraformer-v1":      true,
		"paraformer-8k-v2":  true,
		"paraformer-8k-v1":  true,
		"paraformer-mtl-v1": true,
		"fun-asr":           true,
		"fun-asr-mtl":       true,
	}

	// Audio format to MIME mapping for run-task
	sttAudioFormats = map[string]string{
		".pcm":  "pcm",
		".wav":  "wav",
		".mp3":  "mp3",
		".opus": "opus",
		".spx":  "speex",
		".aac":  "aac",
		".amr":  "amr",
	}
)

// Flag structs
type sttFlags struct {
	file              string
	model             string
	language          string
	noITN             bool
	verbose           bool
	output            string
	vocabularyID      string
	disfluencyRemoval bool
	languageHints     string
	sampleRate        int
}

type sttCreateFlags struct {
	model             string
	languageHints     string
	vocabularyID      string
	disfluencyRemoval bool
	diarize           bool
	speakers          int
	channel           []int
	itn               bool
	words             bool
}

type sttStatusFlags struct {
	verbose bool
	output  string
}

// Commands
var sttCmd = newSTTCmd()

func newSTTCmd() *cobra.Command {
	flags := &sttFlags{}

	cmd := &cobra.Command{
		Use:           "stt [audio-file]",
		Short:         "Speech to Text using DashScope ASR models",
		Long:          "Transcribe audio files using Alibaba DashScope ASR models (Qwen-ASR, Paraformer, Fun-ASR).",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSTTDefault(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "Input audio file path")
	cmd.Flags().StringVarP(&flags.model, "model", "m", "qwen3-asr-flash", "Model name")
	cmd.Flags().StringVarP(&flags.language, "language", "l", "", "Language code (zh, en, ja, etc.)")
	cmd.Flags().BoolVar(&flags.noITN, "no-itn", false, "Disable inverse text normalization")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Include timestamps and segments")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&flags.vocabularyID, "vocabulary-id", "", "Hot words vocabulary ID (realtime only)")
	cmd.Flags().BoolVar(&flags.disfluencyRemoval, "disfluency-removal", false, "Remove filler words (paraformer/fun-asr realtime only)")
	cmd.Flags().StringVar(&flags.languageHints, "language-hints", "", "Comma-separated language hints (paraformer-realtime-v2 only)")
	cmd.Flags().IntVar(&flags.sampleRate, "sample-rate", 0, "Sample rate in Hz (realtime only, auto-detected from file)")

	cmd.AddCommand(newSTTCreateCmd())
	cmd.AddCommand(newSTTStatusCmd())

	return cmd
}

func newSTTCreateCmd() *cobra.Command {
	flags := &sttCreateFlags{}

	cmd := &cobra.Command{
		Use:           "create <url> [url...]",
		Short:         "Submit async transcription task",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSTTCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.model, "model", "m", "paraformer-v2", "Model name")
	cmd.Flags().StringVar(&flags.languageHints, "language-hints", "", "Comma-separated language hints (paraformer-v2 only)")
	cmd.Flags().StringVar(&flags.vocabularyID, "vocabulary-id", "", "Hot words vocabulary ID")
	cmd.Flags().BoolVar(&flags.disfluencyRemoval, "disfluency-removal", false, "Remove filler words (paraformer/fun-asr)")
	cmd.Flags().BoolVar(&flags.diarize, "diarize", false, "Enable speaker diarization")
	cmd.Flags().IntVar(&flags.speakers, "speakers", 0, "Reference speaker count, 2-100 (requires --diarize)")
	cmd.Flags().IntSliceVar(&flags.channel, "channel", []int{0}, "Audio channel indices")
	cmd.Flags().BoolVar(&flags.itn, "itn", false, "Enable ITN (qwen3-asr-flash-filetrans only)")
	cmd.Flags().BoolVar(&flags.words, "words", false, "Enable word-level timestamps (qwen3-asr-flash-filetrans only)")

	return cmd
}

func newSTTStatusCmd() *cobra.Command {
	flags := &sttStatusFlags{}

	cmd := &cobra.Command{
		Use:           "status <task_id>",
		Short:         "Query async transcription task status",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSTTStatus(cmd, args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show full output including per-file details")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Save transcription result(s) to file")

	return cmd
}

// ===== STT Default Command =====

func runSTTDefault(cmd *cobra.Command, args []string, flags *sttFlags) error {
	// Get audio input
	audioFile, cleanup, err := getSTTAudioInput(args, flags.file, cmd.InOrStdin())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			return common.WriteError(cmd, "file_not_found", err.Error())
		}
		return common.WriteError(cmd, "missing_input", err.Error())
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Validate model
	isSync := validSyncSTTModels[flags.model]
	isRunTask := validRunTaskSTTModels[flags.model]
	isSessionUpdate := validSessionUpdateSTTModels[flags.model]

	if !isSync && !isRunTask && !isSessionUpdate {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s' for stt command", flags.model))
	}

	// Validate file size (sync models only)
	if isSync {
		info, statErr := os.Stat(audioFile)
		if statErr != nil {
			return common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot access file: %s", statErr.Error()))
		}
		if info.Size() > sttMaxSyncFileSize {
			return common.WriteError(cmd, "file_too_large", fmt.Sprintf("file size %d bytes exceeds 10 MB limit for sync model, use a realtime model for larger files", info.Size()))
		}
	}

	// Compatibility checks
	if flags.language != "" {
		// --language only works with qwen3-asr-flash and qwen3-asr-flash-realtime
		if isRunTask {
			return common.WriteError(cmd, "incompatible_language", "--language is not supported by this model, use --language-hints instead")
		}
	}

	if flags.languageHints != "" {
		// --language-hints only works with run-task models (paraformer-realtime-v2, fun-asr-realtime)
		if !sttLanguageHintsRealtimeModels[flags.model] {
			return common.WriteError(cmd, "incompatible_language_hints", "--language-hints is not supported by this model, use --language instead")
		}
	}

	if flags.noITN {
		// --no-itn only works with qwen3-asr-flash
		if !isSync {
			return common.WriteError(cmd, "incompatible_no_itn", "--no-itn is only supported by qwen3-asr-flash")
		}
	}

	if flags.vocabularyID != "" {
		if !sttVocabularyModels[flags.model] {
			return common.WriteError(cmd, "incompatible_vocabulary_id", "--vocabulary-id is not supported by this model")
		}
	}

	if flags.disfluencyRemoval {
		if !sttDisfluencyModels[flags.model] {
			return common.WriteError(cmd, "incompatible_disfluency_removal", "--disfluency-removal is not supported by this model")
		}
	}

	// Check API key
	apiKey := config.GetAPIKey("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("DASHSCOPE_API_KEY"))
	}

	// Call appropriate API
	var result map[string]any
	if isSync {
		result, err = runSTTSync(cmd, audioFile, apiKey, flags)
	} else if isRunTask {
		result, err = runSTTRunTask(cmd, audioFile, apiKey, flags)
	} else {
		result, err = runSTTSessionUpdate(cmd, audioFile, apiKey, flags)
	}

	if err != nil {
		return err
	}

	// Write to output file if specified
	if flags.output != "" {
		absPath, pathErr := filepath.Abs(flags.output)
		if pathErr != nil {
			absPath = flags.output
		}
		outputData, _ := json.MarshalIndent(result, "", "  ")
		if writeErr := os.WriteFile(absPath, outputData, 0644); writeErr != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write to output file: %s", writeErr.Error()))
		}
		result["file"] = absPath
	}

	return common.WriteSuccess(cmd, result)
}

// ===== STT Create Command =====

func runSTTCreate(cmd *cobra.Command, args []string, flags *sttCreateFlags) error {
	// Get URLs
	if len(args) == 0 {
		return common.WriteError(cmd, "missing_url", "at least one audio URL is required")
	}

	urls := args
	for _, u := range urls {
		if !isURL(u) {
			return common.WriteError(cmd, "missing_url", fmt.Sprintf("'%s' is not a valid URL, async transcription only supports HTTP/HTTPS URLs", u))
		}
	}

	// Validate model
	if !validAsyncSTTModels[flags.model] {
		return common.WriteError(cmd, "invalid_model", fmt.Sprintf("invalid model '%s' for async transcription", flags.model))
	}

	isQwenFiletrans := strings.HasPrefix(flags.model, "qwen3-asr-flash-filetrans")

	// Compatibility checks
	if flags.languageHints != "" {
		if flags.model != "paraformer-v2" {
			return common.WriteError(cmd, "incompatible_language_hints", "--language-hints is only supported by paraformer-v2")
		}
	}

	if flags.vocabularyID != "" {
		if !sttAsyncVocabularyModels[flags.model] {
			return common.WriteError(cmd, "incompatible_vocabulary_id", "--vocabulary-id is not supported by this model")
		}
	}

	if flags.disfluencyRemoval {
		if !sttAsyncDisfluencyModels[flags.model] {
			return common.WriteError(cmd, "incompatible_disfluency_removal", "--disfluency-removal is not supported by this model")
		}
	}

	if flags.diarize {
		if !sttAsyncDiarizeModels[flags.model] {
			return common.WriteError(cmd, "incompatible_diarize", "--diarize is not supported by this model")
		}
	}

	if flags.itn {
		if !isQwenFiletrans {
			return common.WriteError(cmd, "incompatible_itn", "--itn is only supported by qwen3-asr-flash-filetrans")
		}
	}

	if flags.words {
		if !isQwenFiletrans {
			return common.WriteError(cmd, "incompatible_words", "--words is only supported by qwen3-asr-flash-filetrans")
		}
	}

	if flags.speakers > 0 && !flags.diarize {
		return common.WriteError(cmd, "speakers_requires_diarize", "--speakers requires --diarize")
	}

	if flags.speakers != 0 && (flags.speakers < 2 || flags.speakers > 100) {
		return common.WriteError(cmd, "invalid_speakers", "speaker count must be between 2 and 100")
	}

	// Check API key
	apiKey := config.GetAPIKey("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("DASHSCOPE_API_KEY"))
	}

	// Build request
	input := map[string]any{
		"file_urls": urls,
	}

	params := map[string]any{
		"channel_id": flags.channel,
	}

	if flags.languageHints != "" {
		hints := strings.Split(flags.languageHints, ",")
		for i := range hints {
			hints[i] = strings.TrimSpace(hints[i])
		}
		params["language_hints"] = hints
	}

	if flags.vocabularyID != "" {
		params["vocabulary_id"] = flags.vocabularyID
	}

	if flags.disfluencyRemoval {
		params["disfluency_removal_enabled"] = true
	}

	if flags.diarize {
		params["diarization_enabled"] = true
	}

	if flags.speakers > 0 {
		params["speaker_count"] = flags.speakers
	}

	if flags.itn {
		params["enable_itn"] = true
	}

	if flags.words {
		params["enable_words"] = true
	}

	body := map[string]any{
		"model":      flags.model,
		"input":      input,
		"parameters": params,
	}

	baseURL := getBaseURL()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	req, err := http.NewRequest("POST", baseURL+sttTranscriptionPath, bytes.NewReader(jsonBody))
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
		Code    string `json:"code"`
		Message string `json:"message"`
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

// ===== STT Status Command =====

func runSTTStatus(cmd *cobra.Command, args []string, flags *sttStatusFlags) error {
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

	var taskResult struct {
		Output *struct {
			TaskID     string `json:"task_id"`
			TaskStatus string `json:"task_status"`
			Results    []struct {
				FileURL          string `json:"file_url"`
				TranscriptionURL string `json:"transcription_url"`
				SubtaskStatus    string `json:"subtask_status"`
			} `json:"results"`
			TaskMetrics *struct {
				Total     int `json:"TOTAL"`
				Succeeded int `json:"SUCCEEDED"`
				Failed    int `json:"FAILED"`
			} `json:"task_metrics"`
			Message string `json:"message"`
		} `json:"output"`
		Usage *struct {
			Duration int `json:"duration"`
		} `json:"usage"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(respBody, &taskResult); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if taskResult.Code != "" {
		return common.WriteError(cmd, taskResult.Code, taskResult.Message)
	}

	if taskResult.Output == nil {
		return common.WriteError(cmd, "response_error", "empty response from API")
	}

	status := strings.ToLower(taskResult.Output.TaskStatus)

	if status == "failed" {
		msg := taskResult.Output.Message
		if msg == "" {
			msg = "transcription failed"
		}
		return common.WriteError(cmd, "transcription_failed", msg)
	}

	output := map[string]any{
		"success": true,
		"task_id": taskResult.Output.TaskID,
		"status":  status,
	}

	if status == "succeeded" && taskResult.Output.Results != nil {
		if taskResult.Usage != nil {
			output["duration"] = taskResult.Usage.Duration
		}

		// Download transcription results
		var fileResults []map[string]any
		var combinedText strings.Builder

		for _, r := range taskResult.Output.Results {
			fileResult := map[string]any{
				"file_url": r.FileURL,
				"status":   strings.ToLower(r.SubtaskStatus),
			}

			if strings.ToLower(r.SubtaskStatus) == "succeeded" && r.TranscriptionURL != "" {
				transcript, dlErr := downloadTranscription(r.TranscriptionURL)
				if dlErr == nil && transcript != nil {
					text := extractTranscriptText(transcript)
					fileResult["text"] = text
					if combinedText.Len() > 0 {
						combinedText.WriteString("\n")
					}
					combinedText.WriteString(text)

					if flags.verbose {
						if segments, ok := transcript["transcripts"]; ok {
							fileResult["segments"] = segments
						}
						fileResult["transcription_url"] = r.TranscriptionURL
					}
				}
			} else if strings.ToLower(r.SubtaskStatus) == "failed" {
				fileResult["error"] = "transcription failed for this file"
			}

			fileResults = append(fileResults, fileResult)
		}

		// Single file: flatten to top-level text
		if len(fileResults) == 1 && fileResults[0]["text"] != nil {
			output["text"] = fileResults[0]["text"]
			if flags.verbose {
				output["results"] = fileResults
			}
		} else {
			output["results"] = fileResults
		}

		if flags.verbose && taskResult.Output.TaskMetrics != nil {
			output["metrics"] = map[string]any{
				"total":     taskResult.Output.TaskMetrics.Total,
				"succeeded": taskResult.Output.TaskMetrics.Succeeded,
				"failed":    taskResult.Output.TaskMetrics.Failed,
			}
		}
	}

	// Write to output file if specified
	if flags.output != "" && status == "succeeded" {
		absPath, pathErr := filepath.Abs(flags.output)
		if pathErr != nil {
			absPath = flags.output
		}
		outputData, _ := json.MarshalIndent(output, "", "  ")
		if writeErr := os.WriteFile(absPath, outputData, 0644); writeErr != nil {
			return common.WriteError(cmd, "output_write_error", fmt.Sprintf("cannot write to output file: %s", writeErr.Error()))
		}
		// Replace full results with file reference
		output = map[string]any{
			"success": true,
			"task_id": taskResult.Output.TaskID,
			"status":  status,
			"file":    absPath,
		}
		if taskResult.Usage != nil {
			output["duration"] = taskResult.Usage.Duration
		}
	}

	return common.WriteSuccess(cmd, output)
}

// ===== Sync HTTP Implementation =====

func runSTTSync(cmd *cobra.Command, audioFile, apiKey string, flags *sttFlags) (map[string]any, error) {
	// Read and base64-encode audio
	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		return nil, common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot read audio file: %s", err.Error()))
	}

	ext := strings.ToLower(filepath.Ext(audioFile))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "audio/wav"
	}
	audioBase64 := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(audioData))

	// Build request (OpenAI-compatible protocol)
	messages := []map[string]any{
		{
			"role": "user",
			"content": []map[string]any{
				{
					"type": "input_audio",
					"input_audio": map[string]any{
						"data": audioBase64,
					},
				},
			},
		},
	}

	reqBody := map[string]any{
		"model":    flags.model,
		"messages": messages,
		"stream":   false,
	}

	// Add asr_options if needed
	asrOptions := map[string]any{}
	if flags.language != "" {
		asrOptions["language"] = flags.language
	}
	if flags.noITN {
		asrOptions["enable_itn"] = false
	}
	if len(asrOptions) > 0 {
		reqBody["extra_body"] = map[string]any{
			"asr_options": asrOptions,
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, common.WriteError(cmd, "request_error", fmt.Sprintf("cannot marshal request: %s", err.Error()))
	}

	// Use compatible-mode URL
	baseURL := getBaseURL()
	compatURL := strings.Replace(baseURL, "/api/v1", "/compatible-mode/v1", 1)

	req, err := http.NewRequest("POST", compatURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleHTTPError(cmd, resp.StatusCode, string(respBody))
	}

	// Parse OpenAI-compatible response
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content     string `json:"content"`
				Annotations []struct {
					Type     string `json:"type"`
					Language string `json:"language"`
					Emotion  string `json:"emotion"`
				} `json:"annotations"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if len(chatResp.Choices) == 0 {
		return nil, common.WriteError(cmd, "response_error", "no transcription result in response")
	}

	result := map[string]any{
		"success": true,
		"text":    chatResp.Choices[0].Message.Content,
		"model":   flags.model,
	}

	// Extract language and emotion from annotations
	for _, ann := range chatResp.Choices[0].Message.Annotations {
		if ann.Type == "audio_info" {
			if ann.Language != "" {
				result["language"] = ann.Language
			}
			if ann.Emotion != "" {
				result["emotion"] = ann.Emotion
			}
		}
	}

	return result, nil
}

// ===== WebSocket run-task Implementation =====

func runSTTRunTask(cmd *cobra.Command, audioFile, apiKey string, flags *sttFlags) (map[string]any, error) {
	baseURL := getBaseURL()
	wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "/api/v1", "/api-ws/v1/inference/", 1)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+apiKey)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot connect to WebSocket: %s", err.Error()))
	}
	defer conn.Close()

	// Determine audio format
	ext := strings.ToLower(filepath.Ext(audioFile))
	audioFormat, ok := sttAudioFormats[ext]
	if !ok {
		audioFormat = "wav"
	}

	// Determine sample rate
	sampleRate := flags.sampleRate
	if sampleRate == 0 {
		sampleRate = 16000 // default
	}

	taskID := uuid.New().String()

	// Build parameters
	params := map[string]any{
		"format":      audioFormat,
		"sample_rate": sampleRate,
	}
	if flags.vocabularyID != "" {
		params["vocabulary_id"] = flags.vocabularyID
	}
	if flags.disfluencyRemoval {
		params["disfluency_removal_enabled"] = true
	}
	if flags.languageHints != "" {
		hints := strings.Split(flags.languageHints, ",")
		for i := range hints {
			hints[i] = strings.TrimSpace(hints[i])
		}
		params["language_hints"] = hints
	}

	// Send run-task
	runTask := map[string]any{
		"header": map[string]any{
			"action":    "run-task",
			"task_id":   taskID,
			"streaming": "duplex",
		},
		"payload": map[string]any{
			"task_group": "audio",
			"task":       "asr",
			"function":   "recognition",
			"model":      flags.model,
			"parameters": params,
			"input":      map[string]any{},
		},
	}

	if err := conn.WriteJSON(runTask); err != nil {
		return nil, common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot send run-task: %s", err.Error()))
	}

	// Wait for task-started
	if err := waitForRunTaskEvent(conn, "task-started"); err != nil {
		return nil, common.WriteError(cmd, "websocket_error", fmt.Sprintf("task start failed: %s", err.Error()))
	}

	// Open audio file and stream
	audioFileHandle, err := os.Open(audioFile)
	if err != nil {
		return nil, common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot open audio file: %s", err.Error()))
	}
	defer audioFileHandle.Close()

	// Start reading results in background
	type wsResult struct {
		sentences []map[string]any
		err       error
		duration  float64
	}
	resultCh := make(chan wsResult, 1)

	go func() {
		var sentences []map[string]any
		var totalDuration float64

		for {
			_, message, readErr := conn.ReadMessage()
			if readErr != nil {
				resultCh <- wsResult{sentences: sentences, err: nil, duration: totalDuration}
				return
			}

			var event struct {
				Header struct {
					Event        string `json:"event"`
					ErrorCode    string `json:"error_code"`
					ErrorMessage string `json:"error_message"`
				} `json:"header"`
				Payload struct {
					Output struct {
						Sentence struct {
							BeginTime   *float64 `json:"begin_time"`
							EndTime     *float64 `json:"end_time"`
							Text        string   `json:"text"`
							SentenceEnd bool     `json:"sentence_end"`
							Words       []struct {
								BeginTime   float64 `json:"begin_time"`
								EndTime     float64 `json:"end_time"`
								Text        string  `json:"text"`
								Punctuation string  `json:"punctuation"`
							} `json:"words"`
						} `json:"sentence"`
					} `json:"output"`
					Usage struct {
						Duration float64 `json:"duration"`
					} `json:"usage"`
				} `json:"payload"`
			}

			if jsonErr := json.Unmarshal(message, &event); jsonErr != nil {
				continue
			}

			switch event.Header.Event {
			case "result-generated":
				if event.Payload.Output.Sentence.SentenceEnd {
					sentence := map[string]any{
						"text": event.Payload.Output.Sentence.Text,
					}
					if event.Payload.Output.Sentence.BeginTime != nil {
						sentence["start"] = *event.Payload.Output.Sentence.BeginTime / 1000.0
					}
					if event.Payload.Output.Sentence.EndTime != nil {
						sentence["end"] = *event.Payload.Output.Sentence.EndTime / 1000.0
					}
					if len(event.Payload.Output.Sentence.Words) > 0 {
						var words []map[string]any
						for _, w := range event.Payload.Output.Sentence.Words {
							words = append(words, map[string]any{
								"text":        w.Text,
								"start":       w.BeginTime / 1000.0,
								"end":         w.EndTime / 1000.0,
								"punctuation": w.Punctuation,
							})
						}
						sentence["words"] = words
					}
					sentences = append(sentences, sentence)
					if event.Payload.Usage.Duration > 0 {
						totalDuration = event.Payload.Usage.Duration
					}
				}
			case "task-finished":
				resultCh <- wsResult{sentences: sentences, err: nil, duration: totalDuration}
				return
			case "task-failed":
				resultCh <- wsResult{err: fmt.Errorf("%s: %s", event.Header.ErrorCode, event.Header.ErrorMessage)}
				return
			}
		}
	}()

	// Stream audio data in chunks (~100ms per chunk)
	chunkSize := sampleRate * 2 / 10 // 16-bit mono, 100ms
	buf := make([]byte, chunkSize)
	for {
		n, readErr := audioFileHandle.Read(buf)
		if n > 0 {
			if writeErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
				break
			}
			time.Sleep(10 * time.Millisecond) // Small delay to avoid overwhelming
		}
		if readErr != nil {
			break
		}
	}

	// Send finish-task
	finishTask := map[string]any{
		"header": map[string]any{
			"action":    "finish-task",
			"task_id":   taskID,
			"streaming": "duplex",
		},
		"payload": map[string]any{
			"input": map[string]any{},
		},
	}
	_ = conn.WriteJSON(finishTask)

	// Wait for results
	wsRes := <-resultCh
	if wsRes.err != nil {
		return nil, common.WriteError(cmd, "server_error", fmt.Sprintf("recognition failed: %s", wsRes.err.Error()))
	}

	// Build result
	var textParts []string
	for _, s := range wsRes.sentences {
		if t, ok := s["text"].(string); ok {
			textParts = append(textParts, t)
		}
	}

	result := map[string]any{
		"success": true,
		"text":    strings.Join(textParts, ""),
		"model":   flags.model,
	}

	if wsRes.duration > 0 {
		result["duration"] = wsRes.duration
	}

	if flags.verbose && len(wsRes.sentences) > 0 {
		result["segments"] = wsRes.sentences
	}

	return result, nil
}

// ===== WebSocket session.update Implementation =====

func runSTTSessionUpdate(cmd *cobra.Command, audioFile, apiKey string, flags *sttFlags) (map[string]any, error) {
	baseURL := getBaseURL()
	wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "/api/v1", "/api-ws/v1/realtime", 1)
	wsURL += "?model=" + flags.model

	header := http.Header{}
	header.Set("Authorization", "Bearer "+apiKey)
	header.Set("OpenAI-Beta", "realtime=v1")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot connect to WebSocket: %s", err.Error()))
	}
	defer conn.Close()

	// Wait for session.created
	if err := waitForEvent(conn, "session.created"); err != nil {
		return nil, common.WriteError(cmd, "websocket_error", fmt.Sprintf("session creation failed: %s", err.Error()))
	}

	// Determine sample rate and format
	sampleRate := flags.sampleRate
	if sampleRate == 0 {
		sampleRate = 16000
	}

	// Send session.update
	sessionConfig := map[string]any{
		"modalities":         []string{"text"},
		"input_audio_format": "pcm",
		"sample_rate":        sampleRate,
	}
	if flags.language != "" {
		sessionConfig["language"] = flags.language
	}

	sessionUpdate := map[string]any{
		"type":    "session.update",
		"session": sessionConfig,
	}

	if err := conn.WriteJSON(sessionUpdate); err != nil {
		return nil, common.WriteError(cmd, "websocket_error", fmt.Sprintf("cannot send session update: %s", err.Error()))
	}

	// Wait for session.updated
	if err := waitForEvent(conn, "session.updated"); err != nil {
		return nil, common.WriteError(cmd, "websocket_error", fmt.Sprintf("session update failed: %s", err.Error()))
	}

	// Open audio file
	audioFileHandle, err := os.Open(audioFile)
	if err != nil {
		return nil, common.WriteError(cmd, "file_not_found", fmt.Sprintf("cannot open audio file: %s", err.Error()))
	}
	defer audioFileHandle.Close()

	// Read results in background
	type wsResult struct {
		text    string
		emotion string
		err     error
	}
	resultCh := make(chan wsResult, 1)

	go func() {
		var fullText strings.Builder
		var emotion string

		for {
			_, message, readErr := conn.ReadMessage()
			if readErr != nil {
				resultCh <- wsResult{text: fullText.String(), emotion: emotion}
				return
			}

			var event struct {
				Type       string `json:"type"`
				Message    string `json:"message"`
				Transcript struct {
					Text    string `json:"text"`
					Emotion string `json:"emotion"`
				} `json:"transcript"`
			}

			if jsonErr := json.Unmarshal(message, &event); jsonErr != nil {
				continue
			}

			switch event.Type {
			case "conversation.item.input_audio_transcription.completed":
				if event.Transcript.Text != "" {
					fullText.WriteString(event.Transcript.Text)
				}
				if event.Transcript.Emotion != "" {
					emotion = event.Transcript.Emotion
				}
			case "session.finished":
				resultCh <- wsResult{text: fullText.String(), emotion: emotion}
				return
			case "error":
				resultCh <- wsResult{err: fmt.Errorf("%s", event.Message)}
				return
			}
		}
	}()

	// Send audio chunks as base64
	chunkSize := sampleRate * 2 / 10 // 100ms of 16-bit mono PCM
	buf := make([]byte, chunkSize)
	for {
		n, readErr := audioFileHandle.Read(buf)
		if n > 0 {
			appendMsg := map[string]any{
				"type":  "input_audio_buffer.append",
				"audio": base64.StdEncoding.EncodeToString(buf[:n]),
			}
			if writeErr := conn.WriteJSON(appendMsg); writeErr != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if readErr != nil {
			break
		}
	}

	// Commit and finish
	_ = conn.WriteJSON(map[string]any{"type": "input_audio_buffer.commit"})
	_ = conn.WriteJSON(map[string]any{"type": "session.finish"})

	// Wait for results
	wsRes := <-resultCh
	if wsRes.err != nil {
		return nil, common.WriteError(cmd, "server_error", fmt.Sprintf("recognition failed: %s", wsRes.err.Error()))
	}

	result := map[string]any{
		"success": true,
		"text":    wsRes.text,
		"model":   flags.model,
	}

	if wsRes.emotion != "" {
		result["emotion"] = wsRes.emotion
	}

	return result, nil
}

// ===== Helper Functions =====

func getSTTAudioInput(args []string, fileFlag string, stdin io.Reader) (string, func(), error) {
	// From positional argument
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		filePath := args[0]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return "", nil, fmt.Errorf("audio file '%s' does not exist", filePath)
		}
		return filePath, nil, nil
	}

	// From --file flag
	if fileFlag != "" {
		if _, err := os.Stat(fileFlag); os.IsNotExist(err) {
			return "", nil, fmt.Errorf("audio file '%s' does not exist", fileFlag)
		}
		return fileFlag, nil, nil
	}

	// From stdin
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", nil, fmt.Errorf("no audio file provided, use positional argument, --file flag, or pipe from stdin")
			}
		}

		tmpFile, err := os.CreateTemp("", "stt-stdin-*.wav")
		if err != nil {
			return "", nil, fmt.Errorf("cannot create temp file: %s", err.Error())
		}
		if _, err := io.Copy(tmpFile, stdin); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", nil, fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		tmpFile.Close()

		cleanup := func() {
			os.Remove(tmpFile.Name())
		}
		return tmpFile.Name(), cleanup, nil
	}

	return "", nil, fmt.Errorf("no audio file provided, use positional argument, --file flag, or pipe from stdin")
}

func waitForRunTaskEvent(conn *websocket.Conn, expectedEvent string) error {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("connection closed: %s", err.Error())
		}

		var event struct {
			Header struct {
				Event        string `json:"event"`
				ErrorCode    string `json:"error_code"`
				ErrorMessage string `json:"error_message"`
			} `json:"header"`
		}
		if jsonErr := json.Unmarshal(message, &event); jsonErr != nil {
			continue
		}

		if event.Header.Event == expectedEvent {
			return nil
		}
		if event.Header.Event == "task-failed" {
			return fmt.Errorf("%s: %s", event.Header.ErrorCode, event.Header.ErrorMessage)
		}
	}
}

func downloadTranscription(url string) (map[string]any, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func extractTranscriptText(transcript map[string]any) string {
	transcripts, ok := transcript["transcripts"].([]any)
	if !ok || len(transcripts) == 0 {
		return ""
	}

	var texts []string
	for _, t := range transcripts {
		if channel, ok := t.(map[string]any); ok {
			if text, ok := channel["text"].(string); ok && text != "" {
				texts = append(texts, text)
			}
		}
	}
	return strings.Join(texts, "\n")
}
