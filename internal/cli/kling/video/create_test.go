package video

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== Create Command Tests =====

func TestCreate_MissingPrompt(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create")

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

func TestCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat playing piano")

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

func TestCreate_InvalidMode(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat", "--mode", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid mode")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_mode" {
		t.Errorf("expected error code 'invalid_mode', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidRatio(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat", "--ratio", "4:3")

	if err == nil {
		t.Fatal("expected error for invalid ratio")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_ratio" {
		t.Errorf("expected error code 'invalid_ratio', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"too_short", "2"},
		{"too_long", "11"},
		{"zero", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, err := executeCommand(cmd, "create", "A cat", "--duration", tt.duration)

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

func TestCreate_FirstFrameNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat", "--first-frame", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for first frame not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "frame_not_found" {
		t.Errorf("expected error code 'frame_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_LastFrameRequiresFirst(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create", "A cat", "--last-frame", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error for last frame without first frame")
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
	tmpFile, err := os.CreateTemp("", "first_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create", "A cat", "--first-frame", tmpFile.Name(), "--last-frame", "/nonexistent/last.jpg")

	if cmdErr == nil {
		t.Fatal("expected error for last frame not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "frame_not_found" {
		t.Errorf("expected error code 'frame_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_RefImageNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "<<<image_1>>> walks", "--ref-image", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for ref image not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "ref_image_not_found" {
		t.Errorf("expected error code 'ref_image_not_found', got: %s", errorObj["code"])
	}
}

// Test that URL inputs skip file existence check
func TestCreate_FirstFrameURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "A cat dancing",
		"--first-frame", "https://example.com/image.jpg")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should fail at API key check, not file check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected URL to skip file check, got error: %s", errorObj["code"])
	}
}

func TestCreate_RefImageURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "<<<image_1>>> walks",
		"--ref-image", "https://example.com/character.png")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should fail at API key check, not file check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected URL to skip file check, got error: %s", errorObj["code"])
	}
}

func TestCreate_MixedLocalAndURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	// Create a temp file for local image
	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, execErr := executeCommand(cmd, "create", "<<<image_1>>> in scene",
		"--first-frame", tmpFile.Name(),
		"--ref-image", "https://example.com/character.png")

	if execErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should fail at API key check, accepting both local file and URL
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected mixed inputs to work, got error: %s", errorObj["code"])
	}
}

func TestCreate_ConflictingVideoFlags(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create", "Edit video",
		"--ref-video", "https://example.com/ref.mp4",
		"--base-video", "https://example.com/base.mp4")

	if err == nil {
		t.Fatal("expected error for conflicting video flags")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "conflicting_video_flags" {
		t.Errorf("expected error code 'conflicting_video_flags', got: %s", errorObj["code"])
	}
}

func TestCreate_ValidModes(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	modes := []string{"std", "pro"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, err := executeCommand(cmd, "create", "A cat", "--mode", mode)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected mode '%s' to be valid, got error: %s", mode, errorObj["code"])
			}
		})
	}
}

func TestCreate_ValidRatios(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	ratios := []string{"16:9", "9:16", "1:1"}
	for _, ratio := range ratios {
		t.Run(ratio, func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, err := executeCommand(cmd, "create", "A cat", "--ratio", ratio)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected ratio '%s' to be valid, got error: %s", ratio, errorObj["code"])
			}
		})
	}
}

func TestCreate_ValidDurations(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	durations := []string{"3", "5", "7", "10"}
	for _, duration := range durations {
		t.Run(duration+"s", func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, err := executeCommand(cmd, "create", "A cat", "--duration", duration)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected duration '%s' to be valid, got error: %s", duration, errorObj["code"])
			}
		})
	}
}

func TestCreate_WithFirstFrame(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create", "A cat walking", "--first-frame", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected first frame to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_WithRefImage(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	tmpFile, err := os.CreateTemp("", "ref_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create", "<<<image_1>>> walks", "--ref-image", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected ref image to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_WithRefVideo(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create", "Same style as <<<video_1>>>", "--ref-video", "https://example.com/ref.mp4")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected ref video to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_WithBaseVideo(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create", "Edit <<<video_1>>>", "--base-video", "https://example.com/base.mp4")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected base video to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_FromFile(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("A beautiful sunset over the ocean")
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create", "--prompt-file", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected prompt from file to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	cmd.SetIn(strings.NewReader("A beautiful sunset"))

	_, stderr, err := executeCommand(cmd, "create")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected prompt from stdin to be valid, got error: %s", errorObj["code"])
	}
}

func TestCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()

	if cmd.Flag("mode").DefValue != "pro" {
		t.Errorf("expected default mode 'pro', got: %s", cmd.Flag("mode").DefValue)
	}
	if cmd.Flag("duration").DefValue != "5" {
		t.Errorf("expected default duration '5', got: %s", cmd.Flag("duration").DefValue)
	}
	if cmd.Flag("ratio").DefValue != "16:9" {
		t.Errorf("expected default ratio '16:9', got: %s", cmd.Flag("ratio").DefValue)
	}
	if cmd.Flag("watermark").DefValue != "false" {
		t.Errorf("expected default watermark 'false', got: %s", cmd.Flag("watermark").DefValue)
	}
	if cmd.Flag("ref-exclude-sound").DefValue != "false" {
		t.Errorf("expected default ref-exclude-sound 'false', got: %s", cmd.Flag("ref-exclude-sound").DefValue)
	}
}

func TestCreate_AllFlags(t *testing.T) {
	cmd := newCreateCmd()

	flags := []string{
		"first-frame", "last-frame", "ref-image", "ref-video", "base-video",
		"ref-exclude-sound", "prompt-file", "mode", "duration", "ratio", "watermark",
	}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestCreate_ShortFlags(t *testing.T) {
	cmd := newCreateCmd()

	shortFlags := map[string]string{
		"i": "first-frame",
		"f": "prompt-file",
		"d": "duration",
		"r": "ratio",
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
