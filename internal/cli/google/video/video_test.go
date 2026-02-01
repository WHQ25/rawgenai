package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

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

func TestCreate_InvalidModel(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--model", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_model" {
		t.Errorf("expected error code 'invalid_model', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidAspect(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--aspect", "4:3")

	if err == nil {
		t.Fatal("expected error for invalid aspect ratio")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_aspect" {
		t.Errorf("expected error code 'invalid_aspect', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidResolution(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--resolution", "4K")

	if err == nil {
		t.Fatal("expected error for invalid resolution")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_resolution" {
		t.Errorf("expected error code 'invalid_resolution', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"too short", "2"},
		{"not allowed", "5"},
		{"too long", "10"},
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
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

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

func TestCreate_ValidFlags(t *testing.T) {
	cmd := newCreateCmd()

	flags := []string{"file", "first-frame", "last-frame", "ref", "model", "aspect", "resolution", "duration", "negative", "seed"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()

	if cmd.Flag("model").DefValue != "veo-3.1" {
		t.Errorf("expected default model 'veo-3.1', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("aspect").DefValue != "16:9" {
		t.Errorf("expected default aspect '16:9', got: %s", cmd.Flag("aspect").DefValue)
	}
	if cmd.Flag("resolution").DefValue != "720p" {
		t.Errorf("expected default resolution '720p', got: %s", cmd.Flag("resolution").DefValue)
	}
	if cmd.Flag("duration").DefValue != "8" {
		t.Errorf("expected default duration '8', got: %s", cmd.Flag("duration").DefValue)
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

	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newCreateCmd()
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

func TestCreate_FromFileNotFound(t *testing.T) {
	cmd := newCreateCmd()
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

func TestCreate_FromStdin(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

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

func TestCreate_FirstFrameNotFound(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--first-frame", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for first frame not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "first_frame_not_found" {
		t.Errorf("expected error code 'first_frame_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidImageFormat(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	tmpFile, err := os.CreateTemp("", "video_test_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--first-frame", tmpFile.Name())

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

func TestCreate_ValidAspects(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	aspects := []string{"16:9", "9:16"}

	for _, aspect := range aspects {
		t.Run(aspect, func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "--aspect", aspect)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' (valid aspect), got: %s", errorObj["code"])
			}
		})
	}
}

func TestCreate_ValidResolutions(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	// 720p works with any duration
	t.Run("720p", func(t *testing.T) {
		cmd := newCreateCmd()
		_, stderr, err := executeCommand(cmd, "A cat", "--resolution", "720p")

		if err == nil {
			t.Fatal("expected error (missing api key)")
		}

		var resp map[string]any
		if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
			t.Fatalf("expected JSON error output, got: %s", stderr)
		}

		errorObj := resp["error"].(map[string]any)
		if errorObj["code"] != "missing_api_key" {
			t.Errorf("expected error code 'missing_api_key' (valid resolution), got: %s", errorObj["code"])
		}
	})

	// 1080p and 4k require 8s duration
	for _, res := range []string{"1080p", "4k"} {
		t.Run(res+"_with_8s", func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "--resolution", res, "--duration", "8")

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' (valid resolution), got: %s", errorObj["code"])
			}
		})
	}
}

func TestCreate_ResolutionDurationMismatch(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	tests := []struct {
		resolution string
		duration   string
	}{
		{"1080p", "4"},
		{"1080p", "6"},
		{"4k", "4"},
		{"4k", "6"},
	}

	for _, tt := range tests {
		t.Run(tt.resolution+"_"+tt.duration+"s", func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "--resolution", tt.resolution, "--duration", tt.duration)

			if err == nil {
				t.Fatal("expected error for resolution/duration mismatch")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_resolution_duration" {
				t.Errorf("expected error code 'invalid_resolution_duration', got: %s", errorObj["code"])
			}
		})
	}
}

func TestCreate_ValidDurations(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	durations := []string{"4", "6", "8"}

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

func TestCreate_WithNegativePrompt(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--negative", "blurry, low quality")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (negative prompt valid), got: %s", errorObj["code"])
	}
}

func TestCreate_WithSeed(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--seed", "12345")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (seed valid), got: %s", errorObj["code"])
	}
}

// ============ video status tests ============

func TestStatus_MissingOperationID(t *testing.T) {
	cmd := newStatusCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing operation_id")
	}
}

func TestStatus_MissingAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newStatusCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123")

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

func TestDownload_MissingOperationID(t *testing.T) {
	cmd := newDownloadCmd()
	_, _, err := executeCommand(cmd, "-o", "out.mp4")

	if err == nil {
		t.Fatal("expected error for missing operation_id")
	}
}

func TestDownload_MissingOutput(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123")

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
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "-o", "out.avi")

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

func TestDownload_MissingAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "-o", "out.mp4")

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

func TestDownload_ValidFlags(t *testing.T) {
	cmd := newDownloadCmd()

	if cmd.Flag("output") == nil {
		t.Error("expected --output flag")
	}
}

func TestCreate_EmptyStringPrompt(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "")

	if err == nil {
		t.Fatal("expected error for empty string prompt")
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

func TestCreate_WhitespaceOnlyPrompt(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "   \t\n  ")

	if err == nil {
		t.Fatal("expected error for whitespace-only prompt")
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

func TestCreate_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "video_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name())

	if err == nil {
		t.Fatal("expected error for empty file")
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

func TestCreate_EmptyStdin(t *testing.T) {
	cmd := newCreateCmd()
	cmd.SetIn(strings.NewReader(""))

	_, stderr, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for empty stdin")
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

func TestCreate_Veo31FastModel(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--model", "veo-3.1-fast")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected veo-3.1-fast to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_ValidFirstFrame(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	formats := []string{".jpg", ".jpeg", ".png"}
	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "video_test_*"+ext)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A cat", "--first-frame", tmpFile.Name())

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected first frame format %s to be valid, got error: %s", ext, errorObj["code"])
			}
		})
	}
}

func TestCreate_GoogleAPIKeyFallback(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "test-google-key")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat")

	if err == nil {
		return // Unlikely, but if API accepts test key, that's fine
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] == "missing_api_key" {
		t.Error("GOOGLE_API_KEY should be used as fallback")
	}
}

func TestStatus_GoogleAPIKeyFallback(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "test-google-key")

	cmd := newStatusCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123")

	if err == nil {
		return
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] == "missing_api_key" {
		t.Error("GOOGLE_API_KEY should be used as fallback")
	}
}

func TestDownload_GoogleAPIKeyFallback(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "test-google-key")

	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "-o", "out.mp4")

	if err == nil {
		return
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] == "missing_api_key" {
		t.Error("GOOGLE_API_KEY should be used as fallback")
	}
}

func TestStatus_EmptyOperationID(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	cmd := newStatusCmd()
	_, stderr, err := executeCommand(cmd, "   ")

	if err == nil {
		t.Fatal("expected error for empty operation_id")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_operation_id" {
		t.Errorf("expected error code 'missing_operation_id', got: %s", errorObj["code"])
	}
}

func TestDownload_EmptyOperationID(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "   ", "-o", "out.mp4")

	if err == nil {
		t.Fatal("expected error for empty operation_id")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_operation_id" {
		t.Errorf("expected error code 'missing_operation_id', got: %s", errorObj["code"])
	}
}

func TestDownload_UppercaseMP4(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "-o", "out.MP4")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should accept uppercase .MP4 and reach API key check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected .MP4 to be valid, got error: %s", errorObj["code"])
	}
}

func TestDownload_NoExtension(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "-o", "output")

	if err == nil {
		t.Fatal("expected error for no extension")
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

func TestDownload_MovFormat(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "-o", "out.mov")

	if err == nil {
		t.Fatal("expected error for .mov format")
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

func TestGetImageMimeType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.JPG", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.JPEG", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.PNG", "image/png"},
		{"image.unknown", "application/octet-stream"},
		{"image", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getImageMimeType(tt.path)
			if result != tt.expected {
				t.Errorf("getImageMimeType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestHandleAPIError_InvalidAPIKey(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("401 Unauthorized: invalid key"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_api_key" {
		t.Errorf("expected 'invalid_api_key', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_PermissionDenied(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("403 Forbidden: permission denied"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "permission_denied" {
		t.Errorf("expected 'permission_denied', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_OperationNotFound(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("404 Not Found"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "operation_not_found" {
		t.Errorf("expected 'operation_not_found', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_QuotaExceeded(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("429: quota exceeded"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "quota_exceeded" {
		t.Errorf("expected 'quota_exceeded', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_RateLimit(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("429 Too Many Requests"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "rate_limit" {
		t.Errorf("expected 'rate_limit', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_ContentPolicy(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("content blocked by safety policy"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "content_policy" {
		t.Errorf("expected 'content_policy', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_Timeout(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("request timeout"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "timeout" {
		t.Errorf("expected 'timeout', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_Connection(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("connection refused"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "connection_error" {
		t.Errorf("expected 'connection_error', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_ServerError(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("500 Internal Server Error"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "server_error" {
		t.Errorf("expected 'server_error', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_ServerOverloaded(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("503 Service Unavailable"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "server_overloaded" {
		t.Errorf("expected 'server_overloaded', got: %s", errorObj["code"])
	}
}

func TestHandleAPIError_Generic(t *testing.T) {
	cmd := &cobra.Command{}
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := handleAPIError(cmd, fmt.Errorf("some unknown error"))

	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderrBuf.String())), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderrBuf.String())
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "api_error" {
		t.Errorf("expected 'api_error', got: %s", errorObj["code"])
	}
}

func TestCmd_SubcommandRegistration(t *testing.T) {
	subcommands := []string{"create", "extend", "status", "download"}

	for _, name := range subcommands {
		found := false
		for _, cmd := range Cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand '%s' to be registered", name)
		}
	}
}

// ============ video extend tests ============

func TestExtend_MissingOperationID(t *testing.T) {
	cmd := newExtendCmd()
	_, _, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing operation_id")
	}
}

func TestExtend_MissingPrompt(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	cmd := newExtendCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123")

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

func TestExtend_MissingAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newExtendCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "Continue the video")

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

func TestExtend_InvalidModel(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	cmd := newExtendCmd()
	_, stderr, err := executeCommand(cmd, "operations/generate-videos-abc123", "Continue", "--model", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_model" {
		t.Errorf("expected error code 'invalid_model', got: %s", errorObj["code"])
	}
}

func TestExtend_ValidFlags(t *testing.T) {
	cmd := newExtendCmd()

	flags := []string{"file", "model", "negative"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestCreate_LastFrameRequiresFirstFrame(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	tmpFile, err := os.CreateTemp("", "video_test_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--last-frame", tmpFile.Name())

	if err == nil {
		t.Fatal("expected error for last-frame without first-frame")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "last_frame_requires_first" {
		t.Errorf("expected error code 'last_frame_requires_first', got: %s", errorObj["code"])
	}
}

func TestCreate_LastFrameNotFound(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	tmpFile, err := os.CreateTemp("", "video_test_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--first-frame", tmpFile.Name(), "--last-frame", "/nonexistent/last.jpg")

	if err == nil {
		t.Fatal("expected error for last frame not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "last_frame_not_found" {
		t.Errorf("expected error code 'last_frame_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_ValidFirstAndLastFrame(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	firstFile, err := os.CreateTemp("", "video_first_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(firstFile.Name())
	firstFile.Close()

	lastFile, err := os.CreateTemp("", "video_last_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lastFile.Name())
	lastFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--first-frame", firstFile.Name(), "--last-frame", lastFile.Name())

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected first+last frame to be valid, got error: %s", errorObj["code"])
	}
}

// ============ reference image tests ============

func TestCreate_RefNotFound(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--ref", "/nonexistent/ref.jpg")

	if err == nil {
		t.Fatal("expected error for ref not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "ref_not_found" {
		t.Errorf("expected error code 'ref_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_TooManyRefs(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	// Create 4 temp files
	var tmpFiles []*os.File
	for i := 0; i < 4; i++ {
		f, err := os.CreateTemp("", fmt.Sprintf("video_ref_%d_*.jpg", i))
		if err != nil {
			t.Fatal(err)
		}
		tmpFiles = append(tmpFiles, f)
		defer os.Remove(f.Name())
		f.Close()
	}

	cmd := newCreateCmd()
	args := []string{"A cat"}
	for _, f := range tmpFiles {
		args = append(args, "--ref", f.Name())
	}

	_, stderr, err := executeCommand(cmd, args...)

	if err == nil {
		t.Fatal("expected error for too many ref images")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "too_many_refs" {
		t.Errorf("expected error code 'too_many_refs', got: %s", errorObj["code"])
	}
}

func TestCreate_ConflictingImageOptions_RefWithFirstFrame(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	refFile, err := os.CreateTemp("", "video_ref_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(refFile.Name())
	refFile.Close()

	frameFile, err := os.CreateTemp("", "video_frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--ref", refFile.Name(), "--first-frame", frameFile.Name())

	if err == nil {
		t.Fatal("expected error for conflicting image options")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "conflicting_image_options" {
		t.Errorf("expected error code 'conflicting_image_options', got: %s", errorObj["code"])
	}
}

func TestCreate_ValidRef(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	refFile, err := os.CreateTemp("", "video_ref_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(refFile.Name())
	refFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--ref", refFile.Name())

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected ref to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_ValidMultipleRefs(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	// Create 3 temp files (max allowed)
	var tmpFiles []*os.File
	for i := 0; i < 3; i++ {
		f, err := os.CreateTemp("", fmt.Sprintf("video_ref_%d_*.jpg", i))
		if err != nil {
			t.Fatal(err)
		}
		tmpFiles = append(tmpFiles, f)
		defer os.Remove(f.Name())
		f.Close()
	}

	cmd := newCreateCmd()
	args := []string{"A cat"}
	for _, f := range tmpFiles {
		args = append(args, "--ref", f.Name())
	}

	_, stderr, err := executeCommand(cmd, args...)

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected 3 refs to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_RefInvalidFormat(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	tmpFile, err := os.CreateTemp("", "video_ref_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A cat", "--ref", tmpFile.Name())

	if err == nil {
		t.Fatal("expected error for invalid ref format")
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
