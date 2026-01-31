package openai

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestVideo_MissingPrompt(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "-o", "output.mp4")

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

func TestVideo_MissingOutput(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "A cat playing piano")

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

func TestVideo_InvalidFormat(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.avi")

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

func TestVideo_InvalidSize(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.mp4", "--size", "1920x1080")

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

func TestVideo_InvalidSeconds(t *testing.T) {
	tests := []struct {
		name    string
		seconds string
	}{
		{"too short", "2"},
		{"not allowed", "6"},
		{"too long", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.mp4", "--seconds", tt.seconds)

			if err == nil {
				t.Fatal("expected error for invalid seconds")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_seconds" {
				t.Errorf("expected error code 'invalid_seconds', got: %s", errorObj["code"])
			}
		})
	}
}

func TestVideo_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.mp4")

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

func TestVideo_ValidFlags(t *testing.T) {
	cmd := newVideoCmd()

	flags := []string{"output", "file", "image", "model", "size", "seconds", "no-wait", "timeout"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestVideo_DefaultValues(t *testing.T) {
	cmd := newVideoCmd()

	if cmd.Flag("model").DefValue != "sora-2" {
		t.Errorf("expected default model 'sora-2', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("size").DefValue != "1280x720" {
		t.Errorf("expected default size '1280x720', got: %s", cmd.Flag("size").DefValue)
	}
	if cmd.Flag("seconds").DefValue != "4" {
		t.Errorf("expected default seconds '4', got: %s", cmd.Flag("seconds").DefValue)
	}
	if cmd.Flag("timeout").DefValue != "600" {
		t.Errorf("expected default timeout '600', got: %s", cmd.Flag("timeout").DefValue)
	}
	if cmd.Flag("no-wait").DefValue != "false" {
		t.Errorf("expected default no-wait 'false', got: %s", cmd.Flag("no-wait").DefValue)
	}
}

func TestVideo_FromFile(t *testing.T) {
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

	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name(), "-o", "out.mp4")

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

func TestVideo_FromFileNotFound(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "--file", "/nonexistent/file.txt", "-o", "out.mp4")

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

func TestVideo_FromStdin(t *testing.T) {
	cmd := newVideoCmd()
	cmd.SetIn(strings.NewReader("A cat playing piano"))

	_, stderr, err := executeCommand(cmd, "-o", "out.mp4")

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

func TestVideo_ImageNotFound(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.mp4", "--image", "/nonexistent/image.jpg")

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

func TestVideo_InvalidImageFormat(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	// Create a temp file with invalid extension
	tmpFile, err := os.CreateTemp("", "video_test_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newVideoCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.mp4", "--image", tmpFile.Name())

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

func TestVideo_ValidSizes(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	validSizes := []string{"1280x720", "720x1280", "1792x1024", "1024x1792"}

	for _, size := range validSizes {
		t.Run(size, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.mp4", "--size", size)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			// Should pass size validation and fail on API key
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' (valid size), got: %s", errorObj["code"])
			}
		})
	}
}

func TestVideo_ValidSeconds(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	validSeconds := []string{"4", "8", "12"}

	for _, seconds := range validSeconds {
		t.Run(seconds+"s", func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "-o", "out.mp4", "--seconds", seconds)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			// Should pass seconds validation and fail on API key
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' (valid seconds), got: %s", errorObj["code"])
			}
		})
	}
}
