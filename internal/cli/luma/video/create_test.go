package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCreate_MissingAPIKeyBeforeValidation(t *testing.T) {
	// Luma API allows text-to-video (prompt only) or image-to-video (image only)
	// Test that API key is checked before sending request
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidModel(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt", "--model", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_model" {
		t.Errorf("expected error code 'invalid_model', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidRatio(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt", "--ratio", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid ratio")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_ratio" {
		t.Errorf("expected error code 'invalid_ratio', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidDuration(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt", "--duration", "10s")

	if err == nil {
		t.Fatal("expected error for invalid duration")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_duration" {
		t.Errorf("expected error code 'invalid_duration', got: %s", errorObj["code"])
	}
}

func TestCreate_InvalidResolution(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt", "--resolution", "8k")

	if err == nil {
		t.Fatal("expected error for invalid resolution")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_resolution" {
		t.Errorf("expected error code 'invalid_resolution', got: %s", errorObj["code"])
	}
}

func TestCreate_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestCreate_ImageNotFound(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt", "--image", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for image not found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "image_not_found" {
		t.Errorf("expected error code 'image_not_found', got: %s", errorObj["code"])
	}
}

func TestCreate_AllFlags(t *testing.T) {
	cmd := newCreateCmd()

	expectedFlags := []string{"image", "end-frame", "model", "ratio", "duration", "resolution", "loop", "prompt-file"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestCreate_ShortFlags(t *testing.T) {
	cmd := newCreateCmd()

	shortFlags := map[string]string{
		"i": "image",
		"m": "model",
		"r": "ratio",
		"d": "duration",
		"f": "prompt-file",
	}

	for short, long := range shortFlags {
		flag := cmd.Flags().ShorthandLookup(short)
		if flag == nil {
			t.Errorf("expected short flag '-%s' not found", short)
			continue
		}
		if flag.Name != long {
			t.Errorf("short flag '-%s' maps to '%s', expected '%s'", short, flag.Name, long)
		}
	}
}

func TestCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()

	defaults := map[string]string{
		"model":      "ray-2",
		"ratio":      "16:9",
		"duration":   "5s",
		"resolution": "", // resolution is optional, no default
		"loop":       "false",
	}

	for flag, expected := range defaults {
		f := cmd.Flags().Lookup(flag)
		if f == nil {
			t.Errorf("flag '%s' not found", flag)
			continue
		}
		if f.DefValue != expected {
			t.Errorf("flag '%s' default is '%s', expected '%s'", flag, f.DefValue, expected)
		}
	}
}
