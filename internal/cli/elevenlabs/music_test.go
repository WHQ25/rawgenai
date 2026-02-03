package elevenlabs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeMusicCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()

	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestMusic_MissingPrompt(t *testing.T) {
	cmd := newMusicCmd()
	cmd.SetIn(strings.NewReader(""))
	_, stderr, err := executeMusicCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got '%s'", errorObj["code"])
	}
}

func TestMusic_MissingOutput(t *testing.T) {
	cmd := newMusicCmd()
	_, stderr, err := executeMusicCommand(cmd, "test prompt")

	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got '%s'", errorObj["code"])
	}
}

func TestMusic_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"too_short", "1000"},
		{"too_long", "700000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newMusicCmd()
			_, stderr, err := executeMusicCommand(cmd, "test prompt", "-o", "out.mp3", "-d", tt.duration)

			if err == nil {
				t.Fatal("expected error for invalid duration")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_duration" {
				t.Errorf("expected error code 'invalid_duration', got '%s'", errorObj["code"])
			}
		})
	}
}

func TestMusic_InvalidFormat(t *testing.T) {
	cmd := newMusicCmd()
	_, stderr, err := executeMusicCommand(cmd, "test prompt", "-o", "out.mp3", "-f", "invalid_format")

	if err == nil {
		t.Fatal("expected error for invalid format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_format" {
		t.Errorf("expected error code 'invalid_format', got '%s'", errorObj["code"])
	}
}

func TestMusic_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newMusicCmd()
	_, stderr, err := executeMusicCommand(cmd, "test prompt", "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got '%s'", errorObj["code"])
	}
}

func TestMusic_ValidFlags(t *testing.T) {
	cmd := newMusicCmd()

	flags := []string{"output", "file", "duration", "instrumental", "format", "composition-plan", "respect-durations", "speak"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag '%s' not found", flag)
		}
	}

	shortFlags := map[string]string{
		"o": "output",
		"d": "duration",
		"f": "format",
	}
	for short, long := range shortFlags {
		flag := cmd.Flags().Lookup(long)
		if flag == nil || flag.Shorthand != short {
			t.Errorf("short flag '-%s' not mapped to '--%s'", short, long)
		}
	}
}

func TestMusic_DefaultValues(t *testing.T) {
	cmd := newMusicCmd()

	tests := []struct {
		flag     string
		expected string
	}{
		{"format", "mp3_44100_128"},
		{"respect-durations", "true"},
		{"instrumental", "false"},
		{"speak", "false"},
	}

	for _, tt := range tests {
		flag := cmd.Flags().Lookup(tt.flag)
		if flag == nil {
			t.Errorf("flag '%s' not found", tt.flag)
			continue
		}
		if flag.DefValue != tt.expected {
			t.Errorf("flag '%s' default: expected '%s', got '%s'", tt.flag, tt.expected, flag.DefValue)
		}
	}
}

func TestMusic_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newMusicCmd()
	cmd.SetIn(strings.NewReader("jazz fusion track"))
	_, stderr, err := executeMusicCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing API key), but stdin should be read")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected 'missing_api_key', got '%s'", errorObj["code"])
	}
}
