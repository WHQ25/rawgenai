package video

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// =============================================================================
// Avatar Tests
// =============================================================================

func TestAvatar_MissingImage(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-avatar",
		"--audio", "https://example.com/audio.mp3")

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

func TestAvatar_MissingAudio(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-avatar",
		"-i", tmpFile.Name())

	if cmdErr == nil {
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

func TestAvatar_ConflictingAudio(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-avatar",
		"-i", tmpFile.Name(),
		"--audio", "https://example.com/audio.mp3",
		"--audio-id", "audio_123")

	if cmdErr == nil {
		t.Fatal("expected error for conflicting audio sources")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "conflicting_audio" {
		t.Errorf("expected error code 'conflicting_audio', got: %s", errorObj["code"])
	}
}

func TestAvatar_ImageNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-avatar",
		"-i", "/nonexistent/image.jpg",
		"--audio", "https://example.com/audio.mp3")

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

func TestAvatar_AudioNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-avatar",
		"-i", tmpFile.Name(),
		"--audio", "/nonexistent/audio.mp3")

	if cmdErr == nil {
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

func TestAvatar_InvalidMode(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-avatar",
		"-i", tmpFile.Name(),
		"--audio", "https://example.com/audio.mp3",
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

func TestAvatar_MissingAPIKey(t *testing.T) {
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
	_, stderr, cmdErr := executeCommand(cmd, "create-avatar",
		"-i", tmpFile.Name(),
		"--audio", "https://example.com/audio.mp3")

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

func TestAvatar_ValidWithAudioID(t *testing.T) {
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
	_, stderr, cmdErr := executeCommand(cmd, "create-avatar",
		"-i", tmpFile.Name(),
		"--audio-id", "audio_123")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should fail at API key check, not audio validation
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestAvatar_AllFlags(t *testing.T) {
	cmd := newAvatarCmd()

	flags := []string{"image", "audio", "audio-id", "mode", "watermark", "prompt-file"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestAvatar_ShortFlags(t *testing.T) {
	cmd := newAvatarCmd()

	shortFlags := map[string]string{
		"i": "image",
		"a": "audio",
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

func TestAvatar_DefaultValues(t *testing.T) {
	cmd := newAvatarCmd()

	if cmd.Flag("mode").DefValue != "std" {
		t.Errorf("expected default mode 'std', got: %s", cmd.Flag("mode").DefValue)
	}
	if cmd.Flag("watermark").DefValue != "false" {
		t.Errorf("expected default watermark 'false', got: %s", cmd.Flag("watermark").DefValue)
	}
}
