package video

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== Text2Video (create-from-text) Command Tests =====

func TestText2Video_MissingPrompt(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text")

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

func TestText2Video_InvalidModel(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat", "--model", "invalid-model")

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

func TestText2Video_InvalidCameraControlJSON(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat", "--camera-control", "not-valid-json")

	if err == nil {
		t.Fatal("expected error for invalid camera control JSON")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_camera_control" {
		t.Errorf("expected error code 'invalid_camera_control', got: %s", errorObj["code"])
	}
}

func TestText2Video_ValidCameraControl(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat",
		"--camera-control", `{"type":"simple","config":{"horizontal":5}}`)

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should fail at API key check, not camera control validation
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected camera control to be valid, got error: %s", errorObj["code"])
	}
}

func TestText2Video_AllFlags(t *testing.T) {
	cmd := newText2VideoCmd()

	flags := []string{
		"negative", "model", "mode", "duration", "ratio",
		"cfg-scale", "camera-control", "sound", "watermark", "prompt-file",
	}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestText2Video_DefaultValues(t *testing.T) {
	cmd := newText2VideoCmd()

	if cmd.Flag("model").DefValue != "kling-v1" {
		t.Errorf("expected default model 'kling-v1', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("mode").DefValue != "std" {
		t.Errorf("expected default mode 'std', got: %s", cmd.Flag("mode").DefValue)
	}
	if cmd.Flag("duration").DefValue != "5" {
		t.Errorf("expected default duration '5', got: %s", cmd.Flag("duration").DefValue)
	}
	if cmd.Flag("ratio").DefValue != "16:9" {
		t.Errorf("expected default ratio '16:9', got: %s", cmd.Flag("ratio").DefValue)
	}
}

// ===== Compatibility Check Tests =====

func TestText2Video_CameraControlIncompatibleModel(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat",
		"--model", "kling-v2-6",
		"--camera-control", `{"type":"simple"}`)

	if err == nil {
		t.Fatal("expected error for incompatible camera control")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_camera_control" {
		t.Errorf("expected error code 'incompatible_camera_control', got: %s", errorObj["code"])
	}
}

func TestText2Video_CameraControlIncompatibleMode(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat",
		"--model", "kling-v1",
		"--mode", "pro",
		"--camera-control", `{"type":"simple"}`)

	if err == nil {
		t.Fatal("expected error for incompatible camera control mode")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_camera_control" {
		t.Errorf("expected error code 'incompatible_camera_control', got: %s", errorObj["code"])
	}
}

func TestText2Video_SoundIncompatibleModel(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat",
		"--model", "kling-v1",
		"--sound")

	if err == nil {
		t.Fatal("expected error for incompatible sound")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_sound" {
		t.Errorf("expected error code 'incompatible_sound', got: %s", errorObj["code"])
	}
}

func TestText2Video_CameraControlIncompatibleDuration(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat",
		"--model", "kling-v1",
		"--mode", "std",
		"--duration", "10",
		"--camera-control", `{"type":"simple"}`)

	if err == nil {
		t.Fatal("expected error for incompatible camera control duration")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_camera_control" {
		t.Errorf("expected error code 'incompatible_camera_control', got: %s", errorObj["code"])
	}
}

func TestText2Video_CameraControlCompatible(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat",
		"--model", "kling-v1",
		"--mode", "std",
		"--duration", "5",
		"--camera-control", `{"type":"simple","config":{"horizontal":5}}`)

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected camera control to be compatible, got error: %s", errorObj["code"])
	}
}

func TestText2Video_SoundCompatible(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-text", "A cat",
		"--model", "kling-v2-6",
		"--sound")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected sound to be compatible, got error: %s", errorObj["code"])
	}
}
