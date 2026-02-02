package video

import (
	"bytes"
	"encoding/json"
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

// =====================
// Create command tests
// =====================

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

func TestCreate_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"zero", "0"},
		{"too long", "16"},
		{"negative", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A dancing cat", "-d", tt.duration)

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

func TestCreate_InvalidAspect(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A dancing cat", "-a", "1:1")

	if err == nil {
		t.Fatal("expected error for invalid aspect")
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
	_, stderr, err := executeCommand(cmd, "A dancing cat", "-r", "1080p")

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

func TestCreate_ImageNotFound(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A dancing cat", "-i", "/nonexistent/image.png")

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

func TestCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "A dancing cat")

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

	flags := []string{"prompt-file", "image", "duration", "aspect", "resolution"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()

	if cmd.Flag("duration").DefValue != "5" {
		t.Errorf("expected default duration '5', got: %s", cmd.Flag("duration").DefValue)
	}
	if cmd.Flag("aspect").DefValue != "16:9" {
		t.Errorf("expected default aspect '16:9', got: %s", cmd.Flag("aspect").DefValue)
	}
	if cmd.Flag("resolution").DefValue != "720p" {
		t.Errorf("expected default resolution '720p', got: %s", cmd.Flag("resolution").DefValue)
	}
}

func TestCreate_ShortFlags(t *testing.T) {
	cmd := newCreateCmd()

	shortFlags := map[string]string{
		"i": "image",
		"d": "duration",
		"a": "aspect",
		"r": "resolution",
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

func TestCreate_ValidDurationRange(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	for _, d := range []string{"1", "5", "10", "15"} {
		t.Run("d="+d, func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A dancing cat", "-d", d)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected duration=%s to be valid, got error: %s", d, errorObj["code"])
			}
		})
	}
}

func TestCreate_ValidAspectRatios(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	aspects := []string{"16:9", "9:16"}
	for _, aspect := range aspects {
		t.Run(aspect, func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A dancing cat", "-a", aspect)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected aspect '%s' to be valid, got error: %s", aspect, errorObj["code"])
			}
		})
	}
}

func TestCreate_ValidResolutions(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	resolutions := []string{"720p", "480p"}
	for _, res := range resolutions {
		t.Run(res, func(t *testing.T) {
			cmd := newCreateCmd()
			_, stderr, err := executeCommand(cmd, "A dancing cat", "-r", res)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected resolution '%s' to be valid, got error: %s", res, errorObj["code"])
			}
		})
	}
}

func TestCreate_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("A beautiful sunset")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

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

func TestCreate_WithImage(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "image_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "Animate this", "-i", tmpFile.Name())

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (image validated), got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidImageFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "Animate this", "-i", tmpFile.Name())

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

// =====================
// Edit command tests
// =====================

func TestEdit_MissingPrompt(t *testing.T) {
	cmd := newEditCmd()
	_, stderr, err := executeCommand(cmd, "-v", "https://example.com/video.mp4")

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

func TestEdit_MissingVideo(t *testing.T) {
	cmd := newEditCmd()
	_, stderr, err := executeCommand(cmd, "Make it faster")

	if err == nil {
		t.Fatal("expected error for missing video")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_video" {
		t.Errorf("expected error code 'missing_video', got: %s", errorObj["code"])
	}
}

func TestEdit_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	cmd := newEditCmd()
	_, stderr, err := executeCommand(cmd, "Make it faster", "-v", "https://example.com/video.mp4")

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

func TestEdit_ValidFlags(t *testing.T) {
	cmd := newEditCmd()

	flags := []string{"video", "prompt-file"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestEdit_ShortFlags(t *testing.T) {
	cmd := newEditCmd()

	flag := cmd.Flag("video")
	if flag == nil {
		t.Error("flag --video not found")
	} else if flag.Shorthand != "v" {
		t.Errorf("expected short flag '-v' for '--video', got '-%s'", flag.Shorthand)
	}
}

// =====================
// Status command tests
// =====================

func TestStatus_MissingRequestID(t *testing.T) {
	cmd := newStatusCmd()
	_, stderr, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing request_id")
	}

	// cobra.ExactArgs(1) produces a different error, not our JSON
	if !strings.Contains(stderr, "") && err.Error() != "accepts 1 arg(s), received 0" {
		// Check if it's a JSON error (in case we handle it differently)
		var resp map[string]any
		if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr == nil {
			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_request_id" {
				t.Errorf("expected error code 'missing_request_id', got: %s", errorObj["code"])
			}
		}
	}
}

func TestStatus_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	cmd := newStatusCmd()
	_, stderr, err := executeCommand(cmd, "req_abc123")

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

// =====================
// Download command tests
// =====================

func TestDownload_MissingRequestID(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "-o", "output.mp4")

	if err == nil {
		t.Fatal("expected error for missing request_id")
	}

	// cobra.ExactArgs(1) produces a different error
	if err.Error() != "accepts 1 arg(s), received 0" {
		var resp map[string]any
		if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr == nil {
			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_request_id" {
				t.Errorf("expected error code 'missing_request_id', got: %s", errorObj["code"])
			}
		}
	}
}

func TestDownload_MissingOutput(t *testing.T) {
	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "req_abc123")

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
	_, stderr, err := executeCommand(cmd, "req_abc123", "-o", "output.avi")

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
	common.SetupNoConfigEnv(t)
	t.Setenv("XAI_API_KEY", "")

	cmd := newDownloadCmd()
	_, stderr, err := executeCommand(cmd, "req_abc123", "-o", "output.mp4")

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

func TestDownload_ShortFlags(t *testing.T) {
	cmd := newDownloadCmd()

	flag := cmd.Flag("output")
	if flag == nil {
		t.Error("flag --output not found")
	} else if flag.Shorthand != "o" {
		t.Errorf("expected short flag '-o' for '--output', got '-%s'", flag.Shorthand)
	}
}
