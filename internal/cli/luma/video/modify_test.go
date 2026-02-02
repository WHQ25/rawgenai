package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestModify_MissingVideo(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, _, err := executeCommand(cmd, "modify", "--mode", "flex_1")

	// MarkFlagRequired handles this with Cobra error, not JSON
	if err == nil {
		t.Fatal("expected error for missing video")
	}

	// The error message should mention the required flag
	if !strings.Contains(err.Error(), "video") && !strings.Contains(err.Error(), "required") {
		t.Errorf("expected error about missing video flag, got: %s", err.Error())
	}
}

func TestModify_MissingMode(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, _, err := executeCommand(cmd, "modify", "--video", "https://example.com/video.mp4")

	// MarkFlagRequired handles this with Cobra error, not JSON
	if err == nil {
		t.Fatal("expected error for missing mode")
	}

	// The error message should mention the required flag
	if !strings.Contains(err.Error(), "mode") && !strings.Contains(err.Error(), "required") {
		t.Errorf("expected error about missing mode flag, got: %s", err.Error())
	}
}

func TestModify_InvalidMode(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "modify", "--video", "https://example.com/video.mp4", "--mode", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid mode")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_mode" {
		t.Errorf("expected error code 'invalid_mode', got: %s", errorObj["code"])
	}
}

func TestModify_InvalidModel(t *testing.T) {
	setupNoConfigEnv(t)
	t.Setenv("LUMA_API_KEY", "test-key")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "modify", "--video", "https://example.com/video.mp4", "--mode", "flex_1", "--model", "invalid")

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

func TestModify_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "modify", "--video", "https://example.com/video.mp4", "--mode", "flex_1")

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

func TestModify_AllFlags(t *testing.T) {
	cmd := newModifyCmd()

	expectedFlags := []string{"video", "mode", "model", "first-frame", "prompt-file"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestModify_ShortFlags(t *testing.T) {
	cmd := newModifyCmd()

	shortFlags := map[string]string{
		"v": "video",
		"m": "model",
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

func TestModify_DefaultValues(t *testing.T) {
	cmd := newModifyCmd()

	defaults := map[string]string{
		"model": "ray-2",
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
