package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVideoCreate_T2V_MissingPrompt(t *testing.T) {
	cmd := newTestCmd()
	// Without any image flags, it's t2v mode which requires prompt
	_, stderr, err := executeCommand(cmd, "create")
	if err == nil {
		t.Fatal("expected error for missing prompt")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "missing_prompt" {
		t.Errorf("expected error code missing_prompt, got: %v", errObj["code"])
	}
}

func TestVideoCreate_T2V_InvalidModel(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--model", "invalid", "A cat")
	if err == nil {
		t.Fatal("expected error for invalid model")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_model" {
		t.Errorf("expected error code invalid_model, got: %v", errObj["code"])
	}
}

func TestVideoCreate_I2V_AutoDetect(t *testing.T) {
	// With --first-frame, should auto-detect as i2v
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--first-frame", "nonexistent.png", "A cat")
	if err == nil {
		t.Fatal("expected error (file not found or API key)")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	// Should not be missing_prompt or invalid_type error
	errObj := resp["error"].(map[string]any)
	if errObj["code"] == "missing_prompt" || errObj["code"] == "invalid_type" {
		t.Errorf("unexpected error code: %v", errObj["code"])
	}
}

func TestVideoCreate_FL2V_AutoDetect(t *testing.T) {
	// With both --first-frame and --last-frame, should auto-detect as fl2v
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--first-frame", "first.png", "--last-frame", "last.png")
	if err == nil {
		t.Fatal("expected error (file not found or API key)")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	// Should not be missing_frame error since both frames provided
	errObj := resp["error"].(map[string]any)
	if errObj["code"] == "missing_frame" {
		t.Errorf("unexpected error code missing_frame")
	}
}

func TestVideoCreate_S2V_AutoDetect(t *testing.T) {
	// With --subject, should auto-detect as s2v
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--subject", "face.png", "A girl smiles")
	if err == nil {
		t.Fatal("expected error (file not found or API key)")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	// Should not be missing_subject error since subject provided
	errObj := resp["error"].(map[string]any)
	if errObj["code"] == "missing_subject" {
		t.Errorf("unexpected error code missing_subject")
	}
}

func TestVideoCreate_S2V_NoResolution(t *testing.T) {
	// s2v should reject resolution parameter
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--subject", "face.png", "--resolution", "720P", "A girl")
	if err == nil {
		t.Fatal("expected error for resolution in s2v mode")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_parameter" {
		t.Errorf("expected error code invalid_parameter, got: %v", errObj["code"])
	}
}

func TestVideoCreate_S2V_NoDuration(t *testing.T) {
	// s2v should reject non-default duration
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--subject", "face.png", "--duration", "10", "A girl")
	if err == nil {
		t.Fatal("expected error for duration in s2v mode")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_parameter" {
		t.Errorf("expected error code invalid_parameter, got: %v", errObj["code"])
	}
}

func TestVideoCreate_T2V_InvalidResolution(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "create", "--resolution", "480P", "A cat")
	if err == nil {
		t.Fatal("expected error for invalid resolution")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_resolution" {
		t.Errorf("expected error code invalid_resolution, got: %v", errObj["code"])
	}
}

func TestVideoCreate_I2V_InvalidModel(t *testing.T) {
	cmd := newTestCmd()
	// T2V-01 is not valid for i2v mode
	_, stderr, err := executeCommand(cmd, "create", "--first-frame", "img.png", "--model", "T2V-01", "A cat")
	if err == nil {
		t.Fatal("expected error for invalid i2v model")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_model" {
		t.Errorf("expected error code invalid_model, got: %v", errObj["code"])
	}
}

func TestVideoCreate_AllFlags(t *testing.T) {
	cmd := newCreateCmd()
	flags := cmd.Flags()

	expectedFlags := []string{
		"model", "prompt-file", "duration", "resolution",
		"prompt-optimizer", "fast-pretreatment", "callback-url",
		"first-frame", "last-frame", "subject",
	}

	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}
}

func TestVideoCreate_ShortFlags(t *testing.T) {
	cmd := newCreateCmd()
	flags := cmd.Flags()

	shortFlags := map[string]string{
		"m": "model",
		"d": "duration",
		"r": "resolution",
	}

	for short, long := range shortFlags {
		flag := flags.Lookup(long)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected flag --%s to have shorthand -%s, got -%s", long, short, flag.Shorthand)
		}
	}
}

func TestVideoCreate_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()
	flags := cmd.Flags()

	defaults := map[string]string{
		"duration":         "6",
		"prompt-optimizer": "true",
	}

	for name, expected := range defaults {
		flag := flags.Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
			continue
		}
		if flag.DefValue != expected {
			t.Errorf("expected flag --%s default to be %s, got %s", name, expected, flag.DefValue)
		}
	}
}
