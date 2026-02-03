package voice

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	root.SetOut(stdoutBuf)
	root.SetErr(stderrBuf)
	root.SetArgs(args)

	err = root.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// =============================================================================
// Create Tests
// =============================================================================

func TestVoiceCreate_MissingName(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create",
		"--audio", "https://example.com/audio.mp3")

	if err == nil {
		t.Fatal("expected error for missing name")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_name" {
		t.Errorf("expected error code 'missing_name', got: %s", errorObj["code"])
	}
}

func TestVoiceCreate_InvalidName(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create",
		"ThisNameIsWayTooLongForTheAPI",
		"--audio", "https://example.com/audio.mp3")

	if err == nil {
		t.Fatal("expected error for invalid name")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_name" {
		t.Errorf("expected error code 'invalid_name', got: %s", errorObj["code"])
	}
}

func TestVoiceCreate_MissingAudio(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "MyVoice")

	if err == nil {
		t.Fatal("expected error for missing audio")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_audio" {
		t.Errorf("expected error code 'missing_audio', got: %s", errorObj["code"])
	}
}

func TestVoiceCreate_ConflictingSource(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "MyVoice",
		"--audio", "https://example.com/audio.mp3",
		"--video-id", "video_123")

	if err == nil {
		t.Fatal("expected error for conflicting source")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "conflicting_source" {
		t.Errorf("expected error code 'conflicting_source', got: %s", errorObj["code"])
	}
}

func TestVoiceCreate_AudioNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "MyVoice",
		"--audio", "/nonexistent/audio.mp3")

	if err == nil {
		t.Fatal("expected error for audio not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "audio_not_found" {
		t.Errorf("expected error code 'audio_not_found', got: %s", errorObj["code"])
	}
}

func TestVoiceCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "MyVoice",
		"--audio", "https://example.com/audio.mp3")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestVoiceCreate_ValidWithVideoID(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "MyVoice",
		"--video-id", "video_123")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should fail at API key check, not source validation
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestVoiceCreate_AllFlags(t *testing.T) {
	cmd := newCreateCmd()

	flags := []string{"audio", "video-id"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestVoiceCreate_ShortFlags(t *testing.T) {
	cmd := newCreateCmd()

	shortFlags := map[string]string{
		"a": "audio",
	}

	for short, long := range shortFlags {
		flag := cmd.Flag(long)
		if flag == nil {
			t.Errorf("flag --%s not found", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected short flag '-%s' for '--%s', got '-%s'", short, long, flag.Shorthand)
		}
	}
}

// =============================================================================
// Status Tests
// =============================================================================

func TestVoiceStatus_MissingTaskID(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "status")

	if err == nil {
		t.Fatal("expected error for missing task ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

func TestVoiceStatus_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "status", "task_123")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

// =============================================================================
// List Tests
// =============================================================================

func TestVoiceList_InvalidType(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "list", "--type", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid type")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_type" {
		t.Errorf("expected error code 'invalid_type', got: %s", errorObj["code"])
	}
}

func TestVoiceList_InvalidLimit(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "list", "--limit", "0")

	if err == nil {
		t.Fatal("expected error for invalid limit")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_limit" {
		t.Errorf("expected error code 'invalid_limit', got: %s", errorObj["code"])
	}
}

func TestVoiceList_InvalidPage(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "list", "--page", "0")

	if err == nil {
		t.Fatal("expected error for invalid page")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_page" {
		t.Errorf("expected error code 'invalid_page', got: %s", errorObj["code"])
	}
}

func TestVoiceList_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "list")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestVoiceList_AllFlags(t *testing.T) {
	cmd := newListCmd()

	flags := []string{"type", "limit", "page"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestVoiceList_ShortFlags(t *testing.T) {
	cmd := newListCmd()

	shortFlags := map[string]string{
		"t": "type",
		"l": "limit",
		"p": "page",
	}

	for short, long := range shortFlags {
		flag := cmd.Flag(long)
		if flag == nil {
			t.Errorf("flag --%s not found", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected short flag '-%s' for '--%s', got '-%s'", short, long, flag.Shorthand)
		}
	}
}

func TestVoiceList_DefaultValues(t *testing.T) {
	cmd := newListCmd()

	if cmd.Flag("type").DefValue != "custom" {
		t.Errorf("expected default type 'custom', got: %s", cmd.Flag("type").DefValue)
	}
	if cmd.Flag("limit").DefValue != "30" {
		t.Errorf("expected default limit '30', got: %s", cmd.Flag("limit").DefValue)
	}
	if cmd.Flag("page").DefValue != "1" {
		t.Errorf("expected default page '1', got: %s", cmd.Flag("page").DefValue)
	}
}

func TestVoiceList_ValidTypes(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	types := []string{"custom", "official"}
	for _, voiceType := range types {
		t.Run(voiceType, func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, cmdErr := executeCommand(cmd, "list", "--type", voiceType)

			if cmdErr == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			// Should fail at API key check, not type validation
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' for type '%s', got: %s", voiceType, errorObj["code"])
			}
		})
	}
}

func TestVoiceList_TTSType(t *testing.T) {
	// TTS type doesn't require API key - it returns static voice list
	cmd := NewCmd()
	stdout, _, err := executeCommand(cmd, "list", "--type", "tts")

	if err != nil {
		t.Fatalf("unexpected error for tts type: %v", err)
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}

	if resp["success"] != true {
		t.Error("expected success: true")
	}
	if resp["type"] != "tts" {
		t.Errorf("expected type 'tts', got: %v", resp["type"])
	}

	voices, ok := resp["voices"].([]any)
	if !ok {
		t.Fatal("expected voices array")
	}
	if len(voices) == 0 {
		t.Error("expected non-empty voices list")
	}
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestVoiceDelete_MissingVoiceID(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "delete")

	if err == nil {
		t.Fatal("expected error for missing voice ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_voice_id" {
		t.Errorf("expected error code 'missing_voice_id', got: %s", errorObj["code"])
	}
}

func TestVoiceDelete_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "delete", "voice_123")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/audio.mp3", true},
		{"http://example.com/audio.mp3", true},
		{"/path/to/local/file.mp3", false},
		{"audio.mp3", false},
		{"", false},
	}

	for _, test := range tests {
		result := isURL(test.input)
		if result != test.expected {
			t.Errorf("isURL(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestResolveAudioURL_URL(t *testing.T) {
	url := "https://example.com/audio.mp3"
	result, err := resolveAudioURL(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != url {
		t.Errorf("expected %q, got %q", url, result)
	}
}

func TestResolveAudioURL_LocalFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audio_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test audio content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	result, err := resolveAudioURL(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be base64 encoded
	if result == tmpFile.Name() {
		t.Error("expected base64 encoded content, got original path")
	}
	if len(result) == 0 {
		t.Error("expected non-empty base64 content")
	}
}
