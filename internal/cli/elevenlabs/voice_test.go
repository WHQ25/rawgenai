package elevenlabs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeVoiceCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()

	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestVoiceDesign_MissingDescription(t *testing.T) {
	cmd := newVoiceDesignCmd()
	_, stderr, err := executeVoiceCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for missing description")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_description" {
		t.Errorf("expected error code 'missing_description', got '%s'", errorObj["code"])
	}
}

func TestVoiceDesign_InvalidTextLength(t *testing.T) {
	cmd := newVoiceDesignCmd()
	_, stderr, err := executeVoiceCommand(cmd, "A warm voice", "-o", "out.mp3", "--text", "too short")

	if err == nil {
		t.Fatal("expected error for invalid text length")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_text_length" {
		t.Errorf("expected error code 'invalid_text_length', got '%s'", errorObj["code"])
	}
}

func TestVoiceDesign_InvalidLoudness(t *testing.T) {
	tests := []struct {
		name     string
		loudness string
	}{
		{"too_low", "-1.5"},
		{"too_high", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVoiceDesignCmd()
			_, stderr, err := executeVoiceCommand(cmd, "A warm voice", "-o", "out.mp3", "--loudness", tt.loudness)

			if err == nil {
				t.Fatal("expected error for invalid loudness")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_loudness" {
				t.Errorf("expected error code 'invalid_loudness', got '%s'", errorObj["code"])
			}
		})
	}
}

func TestVoiceDesign_InvalidGuidanceScale(t *testing.T) {
	tests := []struct {
		name  string
		scale string
	}{
		{"too_low", "0.5"},
		{"too_high", "25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVoiceDesignCmd()
			_, stderr, err := executeVoiceCommand(cmd, "A warm voice", "-o", "out.mp3", "--guidance-scale", tt.scale)

			if err == nil {
				t.Fatal("expected error for invalid guidance scale")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_guidance_scale" {
				t.Errorf("expected error code 'invalid_guidance_scale', got '%s'", errorObj["code"])
			}
		})
	}
}

func TestVoiceDesign_InvalidFormat(t *testing.T) {
	cmd := newVoiceDesignCmd()
	_, stderr, err := executeVoiceCommand(cmd, "A warm voice", "-o", "out.mp3", "-f", "invalid")

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

func TestVoiceDesign_IncompatibleReferenceAudio(t *testing.T) {
	cmd := newVoiceDesignCmd()
	// Using default model (v2) with reference audio
	_, stderr, err := executeVoiceCommand(cmd, "A warm voice", "-o", "out.mp3", "--reference-audio", "sample.mp3")

	if err == nil {
		t.Fatal("expected error for incompatible reference audio")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_reference_audio" {
		t.Errorf("expected error code 'incompatible_reference_audio', got '%s'", errorObj["code"])
	}
}

func TestVoiceDesign_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newVoiceDesignCmd()
	_, stderr, err := executeVoiceCommand(cmd, "A warm voice", "-o", "out.mp3")

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

func TestVoiceDesign_ValidFlags(t *testing.T) {
	cmd := newVoiceDesignCmd()

	flags := []string{"output", "description", "text", "auto-text", "model", "format", "loudness", "guidance-scale", "seed", "stream-previews", "enhance", "reference-audio", "prompt-strength", "speak"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag '%s' not found", flag)
		}
	}

	shortFlags := map[string]string{
		"o": "output",
		"m": "model",
		"f": "format",
	}
	for short, long := range shortFlags {
		flag := cmd.Flags().Lookup(long)
		if flag == nil || flag.Shorthand != short {
			t.Errorf("short flag '-%s' not mapped to '--%s'", short, long)
		}
	}
}

func TestVoiceDesign_DefaultValues(t *testing.T) {
	cmd := newVoiceDesignCmd()

	tests := []struct {
		flag     string
		expected string
	}{
		{"model", "eleven_multilingual_ttv_v2"},
		{"format", "mp3_44100_128"},
		{"loudness", "0.5"},
		{"guidance-scale", "5"},
		{"prompt-strength", "0.5"},
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

func TestVoiceCreate_MissingVoiceID(t *testing.T) {
	cmd := newVoiceCreateCmd()
	_, _, err := executeVoiceCommand(cmd, "-n", "name", "-d", "description")

	if err == nil {
		t.Fatal("expected error for missing voice ID")
	}
}

func TestVoiceCreate_InvalidLabels(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newVoiceCreateCmd()
	_, stderr, err := executeVoiceCommand(cmd, "-n", "name", "-d", "description", "--voice-id", "abc123", "--labels", "not json")

	if err == nil {
		t.Fatal("expected error for invalid labels")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_labels" {
		t.Errorf("expected error code 'invalid_labels', got '%s'", errorObj["code"])
	}
}

func TestVoiceCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newVoiceCreateCmd()
	_, stderr, err := executeVoiceCommand(cmd, "-n", "name", "-d", "description", "--voice-id", "abc123")

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

func TestVoiceCreate_ValidFlags(t *testing.T) {
	cmd := newVoiceCreateCmd()

	flags := []string{"name", "description", "voice-id", "labels"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag '%s' not found", flag)
		}
	}

	shortFlags := map[string]string{
		"n": "name",
		"d": "description",
	}
	for short, long := range shortFlags {
		flag := cmd.Flags().Lookup(long)
		if flag == nil || flag.Shorthand != short {
			t.Errorf("short flag '-%s' not mapped to '--%s'", short, long)
		}
	}
}

func TestVoicePreview_MissingVoiceID(t *testing.T) {
	cmd := newVoicePreviewCmd()
	_, stderr, err := executeVoiceCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for missing voice ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_voice_id" {
		t.Errorf("expected error code 'missing_voice_id', got '%s'", errorObj["code"])
	}
}

func TestVoicePreview_MissingOutput(t *testing.T) {
	cmd := newVoicePreviewCmd()
	_, stderr, err := executeVoiceCommand(cmd, "abc123")

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

func TestVoicePreview_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newVoicePreviewCmd()
	_, stderr, err := executeVoiceCommand(cmd, "abc123", "-o", "out.mp3")

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

func TestVoicePreview_ValidFlags(t *testing.T) {
	cmd := newVoicePreviewCmd()

	flags := []string{"output", "speak"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag '%s' not found", flag)
		}
	}

	if flag := cmd.Flags().Lookup("output"); flag.Shorthand != "o" {
		t.Error("short flag '-o' not mapped to '--output'")
	}
}
