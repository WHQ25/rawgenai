package video

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err = cmd.Execute()

	return stdoutBuf.String(), stderrBuf.String(), err
}

// ============ video create tests ============

func TestCreate_MissingPrompt(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidSize(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--size", "1920x1080")

	if err == nil {
		t.Fatal("expected error for invalid size")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_size" {
		t.Errorf("expected error code 'invalid_size', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"too short", "2"},
		{"not allowed", "6"},
		{"too long", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "--duration", tt.duration)

			if err == nil {
				t.Fatal("expected error for invalid duration")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_duration" {
				t.Errorf("expected error code 'invalid_duration', got: %s", errorObj["code"])
			}
		})
	}
}

func TestCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat")

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

func TestCreate_APIKeyFromConfig(t *testing.T) {
	var receivedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"video_123","status":"queued"}`))
	}))
	defer server.Close()

	common.SetupConfigWithAPIKey(t, map[string]string{"openai_api_key": "sk-test-config-key"})
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", server.URL)

	cmd := newCreateCmd()
	executeCommand(cmd, "A cat")

	if receivedKey != "sk-test-config-key" {
		t.Errorf("expected client to receive API key from config, got: %q", receivedKey)
	}
}

func TestCreate_ValidFlags(t *testing.T) {
	cmd := newCreateCmd()

	flags := []string{"prompt-file", "image", "model", "size", "duration"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()

	if cmd.Flag("model").DefValue != "sora-2" {
		t.Errorf("expected default model 'sora-2', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("size").DefValue != "1280x720" {
		t.Errorf("expected default size '1280x720', got: %s", cmd.Flag("size").DefValue)
	}
	if cmd.Flag("duration").DefValue != "4" {
		t.Errorf("expected default duration '4', got: %s", cmd.Flag("duration").DefValue)
	}
}

func TestCreate_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "video_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("A cat playing piano")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", tmpFile.Name())

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (file read success), got: %s", errorObj["code"])
	}
}

func TestCreate_FromFileNotFound(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", "/nonexistent/file.txt")

	if err == nil {
		t.Fatal("expected error for file not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestCreate_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newCreateCmd()
	cmd.SetIn(strings.NewReader("A cat playing piano"))

	_, stderr, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (stdin read success), got: %s", errorObj["code"])
	}
}

func TestCreate_ImageNotFound(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--image", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for image not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "image_not_found" {
		t.Errorf("expected error code 'image_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidImageFormat(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	tmpFile, err := os.CreateTemp("", "video_test_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--image", tmpFile.Name())

	if err == nil {
		t.Fatal("expected error for invalid image format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_image_format" {
		t.Errorf("expected error code 'invalid_image_format', got: %s", errorObj["code"])
	}
}

func TestCreate_ValidSizes(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	sizes := []string{"1280x720", "720x1280", "1792x1024", "1024x1792"}

	for _, size := range sizes {
		t.Run(size, func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "--size", size)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' (valid size), got: %s", errorObj["code"])
			}
		})
	}
}

func TestCreate_ValidDurations(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	durations := []string{"4", "8", "12"}

	for _, duration := range durations {
		t.Run(duration+"s", func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "--duration", duration)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' (valid duration), got: %s", errorObj["code"])
			}
		})
	}
}

// ============ video status tests ============

func TestStatus_MissingVideoID(t *testing.T) {
	cmd := newStatusCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestStatus_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newStatusCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123")

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

func TestStatus_APIKeyFromConfig(t *testing.T) {
	var receivedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"video_abc123","status":"completed"}`))
	}))
	defer server.Close()

	common.SetupConfigWithAPIKey(t, map[string]string{"openai_api_key": "sk-test-config-key"})
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", server.URL)

	cmd := newStatusCmd()
	executeCommand(cmd, "video_abc123")

	if receivedKey != "sk-test-config-key" {
		t.Errorf("expected client to receive API key from config, got: %q", receivedKey)
	}
}

// ============ video download tests ============

func TestDownload_MissingVideoID(t *testing.T) {
	cmd := newDownloadCmd()
	_, _, err := executeCommand(cmd, "-o", "out.mp4")

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestDownload_MissingOutput(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123")

	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestDownload_InvalidFormat(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123", "-o", "out.avi")

	if err == nil {
		t.Fatal("expected error for invalid format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_format" {
		t.Errorf("expected error code 'invalid_format', got: %s", errorObj["code"])
	}
}

func TestDownload_InvalidVariant(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123", "-o", "out.mp4", "--variant", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid variant")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_variant" {
		t.Errorf("expected error code 'invalid_variant', got: %s", errorObj["code"])
	}
}

func TestDownload_ThumbnailRequiresJpg(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123", "-o", "out.mp4", "--variant", "thumbnail")

	if err == nil {
		t.Fatal("expected error for wrong extension")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_format" {
		t.Errorf("expected error code 'invalid_format', got: %s", errorObj["code"])
	}
}

func TestDownload_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123", "-o", "out.mp4")

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

func TestDownload_APIKeyFromConfig(t *testing.T) {
	var receivedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"video_abc123","status":"completed","downloads":{"video":{"url":"http://example.com/video.mp4"}}}`))
	}))
	defer server.Close()

	common.SetupConfigWithAPIKey(t, map[string]string{"openai_api_key": "sk-test-config-key"})
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", server.URL)

	tmpDir := t.TempDir()
	cmd := newDownloadCmd()
	executeCommand(cmd, "video_abc123", "-o", tmpDir+"/out.mp4")

	if receivedKey != "sk-test-config-key" {
		t.Errorf("expected client to receive API key from config, got: %q", receivedKey)
	}
}

func TestDownload_ValidFlags(t *testing.T) {
	cmd := newDownloadCmd()

	if cmd.Flag("output") == nil {
		t.Error("expected --output flag")
	}
	if cmd.Flag("variant") == nil {
		t.Error("expected --variant flag")
	}
}

// ============ video list tests ============

func TestList_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newListCmd()
	_, stderr, err := executeCommand(cmd)

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

func TestList_APIKeyFromConfig(t *testing.T) {
	var receivedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	common.SetupConfigWithAPIKey(t, map[string]string{"openai_api_key": "sk-test-config-key"})
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", server.URL)

	cmd := newListCmd()
	executeCommand(cmd)

	if receivedKey != "sk-test-config-key" {
		t.Errorf("expected client to receive API key from config, got: %q", receivedKey)
	}
}

func TestList_InvalidOrder(t *testing.T) {
	cmd := newListCmd()
	_, stderr, err := executeCommand(cmd, "--order", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid order")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_order" {
		t.Errorf("expected error code 'invalid_order', got: %s", errorObj["code"])
	}
}

func TestList_ValidFlags(t *testing.T) {
	cmd := newListCmd()

	if cmd.Flag("limit") == nil {
		t.Error("expected --limit flag")
	}
	if cmd.Flag("order") == nil {
		t.Error("expected --order flag")
	}
}

// ============ video delete tests ============

func TestDelete_MissingVideoID(t *testing.T) {
	cmd := newDeleteCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestDelete_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newDeleteCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123")

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

func TestDelete_APIKeyFromConfig(t *testing.T) {
	var receivedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	common.SetupConfigWithAPIKey(t, map[string]string{"openai_api_key": "sk-test-config-key"})
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", server.URL)

	cmd := newDeleteCmd()
	executeCommand(cmd, "video_abc123")

	if receivedKey != "sk-test-config-key" {
		t.Errorf("expected client to receive API key from config, got: %q", receivedKey)
	}
}

// ============ video remix tests ============

func TestRemix_MissingVideoID(t *testing.T) {
	cmd := newRemixCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestRemix_MissingPrompt(t *testing.T) {
	cmd := newRemixCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123")

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestRemix_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newRemixCmd()
	_, stderr, err := executeCommand(cmd, "video_abc123", "New prompt")

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

func TestRemix_APIKeyFromConfig(t *testing.T) {
	var receivedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"video_456","status":"queued"}`))
	}))
	defer server.Close()

	common.SetupConfigWithAPIKey(t, map[string]string{"openai_api_key": "sk-test-config-key"})
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", server.URL)

	cmd := newRemixCmd()
	executeCommand(cmd, "video_abc123", "New prompt")

	if receivedKey != "sk-test-config-key" {
		t.Errorf("expected client to receive API key from config, got: %q", receivedKey)
	}
}
