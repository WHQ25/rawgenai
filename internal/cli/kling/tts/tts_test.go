package tts

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	root.SetOut(stdoutBuf)
	root.SetErr(stderrBuf)
	root.SetArgs(args)

	err = root.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestTTS_MissingText(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd)
	if err == nil {
		t.Fatal("expected error for missing text")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_text" {
		t.Errorf("expected error code 'missing_text', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingVoice(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello")
	if err == nil {
		t.Fatal("expected error for missing voice")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_voice" {
		t.Errorf("expected error code 'missing_voice', got: %s", errorObj["code"])
	}
}

func TestTTS_InvalidLanguage(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "--voice", "voice_123", "--language", "jp")
	if err == nil {
		t.Fatal("expected error for invalid language")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_language" {
		t.Errorf("expected error code 'invalid_language', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingOutput(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "--voice", "voice_123")
	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "--voice", "voice_123", "-o", "out.mp3")
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

func TestTTS_InvalidSpeed(t *testing.T) {
	tests := []struct {
		name  string
		speed string
	}{
		{"too_low", "0.5"},
		{"too_high", "3.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "--voice", "voice_123", "-o", "out.mp3", "--speed", tt.speed)
			if err == nil {
				t.Fatal("expected error for invalid speed")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_speed" {
				t.Errorf("expected error code 'invalid_speed', got: %s", errorObj["code"])
			}
		})
	}
}

func TestTTS_AllFlags(t *testing.T) {
	cmd := newTTSCmd()

	expectedFlags := []string{
		"output",
		"prompt-file",
		"voice",
		"language",
		"speed",
		"speak",
	}

	for _, name := range expectedFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s to be registered", name)
		}
	}
}

func TestTTS_ShortFlags(t *testing.T) {
	cmd := newTTSCmd()

	shortFlags := map[string]string{
		"o": "output",
		"f": "prompt-file",
	}

	for short, long := range shortFlags {
		flag := cmd.Flags().Lookup(long)
		if flag == nil {
			t.Errorf("flag --%s not found", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected --%s to have shorthand -%s, got -%s", long, short, flag.Shorthand)
		}
	}
}

func TestTTS_DefaultValues(t *testing.T) {
	cmd := newTTSCmd()

	defaults := map[string]string{
		"language": "zh",
		"speed":    "1",
		"speak":    "false",
	}

	for name, expected := range defaults {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("flag --%s not found", name)
			continue
		}
		if flag.DefValue != expected {
			t.Errorf("expected --%s default to be %q, got %q", name, expected, flag.DefValue)
		}
	}
}

func TestTTS_SpeakWithoutOutput(t *testing.T) {
	// --speak should allow missing -o flag
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "--voice", "voice_123", "--speak")
	if err == nil {
		t.Fatal("expected error (missing_api_key, not missing_output)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	// Should reach API key check, not output check
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}
