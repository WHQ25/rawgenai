package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestText2Video_MissingPrompt(t *testing.T) {
	cmd := newTestCmd()
	cmd.AddCommand(newText2VideoCmd())
	_, stderr, err := executeCommand(cmd, "text2video")

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestText2Video_InvalidModel(t *testing.T) {
	cmd := newTestCmd()
	cmd.AddCommand(newText2VideoCmd())
	_, stderr, err := executeCommand(cmd, "text2video", "test prompt", "-m", "invalid")

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

func TestText2Video_InvalidDuration(t *testing.T) {
	cmd := newTestCmd()
	cmd.AddCommand(newText2VideoCmd())
	_, stderr, err := executeCommand(cmd, "text2video", "test prompt", "-d", "5")

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

func TestText2Video_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	cmd.AddCommand(newText2VideoCmd())
	_, stderr, err := executeCommand(cmd, "text2video", "test prompt")

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

func TestText2Video_AllFlags(t *testing.T) {
	cmd := newText2VideoCmd()

	expectedFlags := []string{"model", "ratio", "duration", "audio", "prompt-file"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestText2Video_DefaultValues(t *testing.T) {
	cmd := newText2VideoCmd()

	defaults := map[string]string{
		"model":    "veo3.1",
		"ratio":    "1280:720",
		"duration": "4",
		"audio":    "true",
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
