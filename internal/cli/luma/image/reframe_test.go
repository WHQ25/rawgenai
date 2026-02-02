package image

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReframe_MissingImage(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, _, err := executeCommand(cmd, "reframe")

	// MarkFlagRequired handles this with Cobra error, not JSON
	if err == nil {
		t.Fatal("expected error for missing image")
	}

	// The error message should mention the required flag
	if !strings.Contains(err.Error(), "image") && !strings.Contains(err.Error(), "required") {
		t.Errorf("expected error about missing image flag, got: %s", err.Error())
	}
}

func TestReframe_ImageNotFound(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "reframe", "--image", "/nonexistent/image.jpg")

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

func TestReframe_InvalidModel(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "reframe", "--image", "https://example.com/image.jpg", "--model", "invalid")

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

func TestReframe_InvalidRatio(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "reframe", "--image", "https://example.com/image.jpg", "--ratio", "invalid")

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

func TestReframe_InvalidFormat(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "reframe", "--image", "https://example.com/image.jpg", "--format", "gif")

	if err == nil {
		t.Fatal("expected error for invalid format")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_format" {
		t.Errorf("expected error code 'invalid_format', got: %s", errorObj["code"])
	}
}

func TestReframe_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "reframe", "--image", "https://example.com/image.jpg")

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

func TestReframe_AllFlags(t *testing.T) {
	cmd := newReframeCmd()

	expectedFlags := []string{"image", "model", "ratio", "format", "prompt-file"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestReframe_ShortFlags(t *testing.T) {
	cmd := newReframeCmd()

	shortFlags := map[string]string{
		"i": "image",
		"m": "model",
		"r": "ratio",
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

func TestReframe_DefaultValues(t *testing.T) {
	cmd := newReframeCmd()

	defaults := map[string]string{
		"model":  "photon-1",
		"ratio":  "16:9",
		"format": "jpg",
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
