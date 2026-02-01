package openai

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// ============ video create tests ============

func TestVideoCreate_MissingPrompt(t *testing.T) {
	cmd := newVideoCreateCmd()
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

func TestVideoCreate_InvalidSize(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoCreateCmd()
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

func TestVideoCreate_InvalidDuration(t *testing.T) {
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
			cmd := newVideoCreateCmd()
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

func TestVideoCreate_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoCreateCmd()
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

func TestVideoCreate_ValidFlags(t *testing.T) {
	cmd := newVideoCreateCmd()

	flags := []string{"file", "image", "model", "size", "duration"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestVideoCreate_DefaultValues(t *testing.T) {
	cmd := newVideoCreateCmd()

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

func TestVideoCreate_FromFile(t *testing.T) {
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

	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoCreateCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name())

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

func TestVideoCreate_FromFileNotFound(t *testing.T) {
	cmd := newVideoCreateCmd()
	_, stderr, err := executeCommand(cmd, "--file", "/nonexistent/file.txt")

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

func TestVideoCreate_FromStdin(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoCreateCmd()
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

func TestVideoCreate_ImageNotFound(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	cmd := newVideoCreateCmd()
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

func TestVideoCreate_InvalidImageFormat(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	tmpFile, err := os.CreateTemp("", "video_test_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newVideoCreateCmd()
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

func TestVideoCreate_ValidSizes(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	validSizes := []string{"1280x720", "720x1280", "1792x1024", "1024x1792"}

	for _, size := range validSizes {
		t.Run(size, func(t *testing.T) {
			cmd := newVideoCreateCmd()
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

func TestVideoCreate_ValidDurations(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	validDurations := []string{"4", "8", "12"}

	for _, duration := range validDurations {
		t.Run(duration+"s", func(t *testing.T) {
			cmd := newVideoCreateCmd()
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

func TestVideoStatus_MissingVideoID(t *testing.T) {
	cmd := newVideoStatusCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestVideoStatus_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoStatusCmd()
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

// ============ video download tests ============

func TestVideoDownload_MissingVideoID(t *testing.T) {
	cmd := newVideoDownloadCmd()
	_, _, err := executeCommand(cmd, "-o", "out.mp4")

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestVideoDownload_MissingOutput(t *testing.T) {
	cmd := newVideoDownloadCmd()
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

func TestVideoDownload_InvalidFormat(t *testing.T) {
	cmd := newVideoDownloadCmd()
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

func TestVideoDownload_InvalidVariant(t *testing.T) {
	cmd := newVideoDownloadCmd()
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

func TestVideoDownload_ThumbnailRequiresJpg(t *testing.T) {
	cmd := newVideoDownloadCmd()
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

func TestVideoDownload_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoDownloadCmd()
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

func TestVideoDownload_ValidFlags(t *testing.T) {
	cmd := newVideoDownloadCmd()

	if cmd.Flag("output") == nil {
		t.Error("expected --output flag")
	}
	if cmd.Flag("variant") == nil {
		t.Error("expected --variant flag")
	}
}

// ============ video list tests ============

func TestVideoList_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoListCmd()
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

func TestVideoList_InvalidOrder(t *testing.T) {
	cmd := newVideoListCmd()
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

func TestVideoList_ValidFlags(t *testing.T) {
	cmd := newVideoListCmd()

	if cmd.Flag("limit") == nil {
		t.Error("expected --limit flag")
	}
	if cmd.Flag("order") == nil {
		t.Error("expected --order flag")
	}
}

// ============ video delete tests ============

func TestVideoDelete_MissingVideoID(t *testing.T) {
	cmd := newVideoDeleteCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestVideoDelete_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoDeleteCmd()
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

// ============ video remix tests ============

func TestVideoRemix_MissingVideoID(t *testing.T) {
	cmd := newVideoRemixCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing video_id")
	}
}

func TestVideoRemix_MissingPrompt(t *testing.T) {
	cmd := newVideoRemixCmd()
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

func TestVideoRemix_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoRemixCmd()
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
