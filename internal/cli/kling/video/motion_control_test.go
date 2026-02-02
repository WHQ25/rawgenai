package video

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== Motion Control Command Tests =====

func TestMotionControl_MissingImage(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-motion-control",
		"--video", "https://example.com/video.mp4")

	if err == nil {
		t.Fatal("expected error for missing image")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_image" {
		t.Errorf("expected error code 'missing_image', got: %s", errorObj["code"])
	}
}

func TestMotionControl_MissingVideo(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-motion-control",
		"-i", tmpFile.Name())

	if cmdErr == nil {
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

func TestMotionControl_ImageNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-motion-control",
		"-i", "/nonexistent/image.jpg",
		"-v", "https://example.com/video.mp4")

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

func TestMotionControl_InvalidOrientation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-motion-control",
		"-i", tmpFile.Name(),
		"-v", "https://example.com/video.mp4",
		"-o", "invalid")

	if cmdErr == nil {
		t.Fatal("expected error for invalid orientation")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_orientation" {
		t.Errorf("expected error code 'invalid_orientation', got: %s", errorObj["code"])
	}
}

func TestMotionControl_InvalidMode(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-motion-control",
		"-i", tmpFile.Name(),
		"-v", "https://example.com/video.mp4",
		"-m", "invalid")

	if cmdErr == nil {
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

func TestMotionControl_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-motion-control",
		"-i", tmpFile.Name(),
		"-v", "https://example.com/video.mp4")

	if cmdErr == nil {
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

func TestMotionControl_ValidOrientations(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	orientations := []string{"image", "video"}
	for _, orientation := range orientations {
		t.Run(orientation, func(t *testing.T) {
			cmd := NewCmd()
			_, stderr, cmdErr := executeCommand(cmd, "create-motion-control",
				"-i", tmpFile.Name(),
				"-v", "https://example.com/video.mp4",
				"-o", orientation)

			if cmdErr == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected orientation '%s' to be valid, got error: %s", orientation, errorObj["code"])
			}
		})
	}
}

func TestMotionControl_AllFlags(t *testing.T) {
	cmd := newMotionControlCmd()

	flags := []string{"image", "video", "orientation", "mode", "keep-sound", "watermark", "prompt-file"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestMotionControl_ShortFlags(t *testing.T) {
	cmd := newMotionControlCmd()

	shortFlags := map[string]string{
		"i": "image",
		"v": "video",
		"o": "orientation",
		"m": "mode",
		"f": "prompt-file",
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

func TestMotionControl_DefaultValues(t *testing.T) {
	cmd := newMotionControlCmd()

	if cmd.Flag("orientation").DefValue != "image" {
		t.Errorf("expected default orientation 'image', got: %s", cmd.Flag("orientation").DefValue)
	}
	if cmd.Flag("mode").DefValue != "std" {
		t.Errorf("expected default mode 'std', got: %s", cmd.Flag("mode").DefValue)
	}
	if cmd.Flag("keep-sound").DefValue != "true" {
		t.Errorf("expected default keep-sound 'true', got: %s", cmd.Flag("keep-sound").DefValue)
	}
}
