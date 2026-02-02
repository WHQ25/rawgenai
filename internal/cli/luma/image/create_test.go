package image

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCreate_MissingPrompt(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
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

func TestCreate_InvalidFormat(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt", "--format", "gif")

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

func TestCreate_ImageRefNotFound(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "test prompt", "--image-ref", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for image ref not found")
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

	expectedFlags := []string{"model", "ratio", "format", "image-ref", "style-ref", "modify-ref", "prompt-file"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestCreate_ShortFlags(t *testing.T) {
	cmd := newCreateCmd()

	shortFlags := map[string]string{
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

func TestCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()

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
