package elevenlabs

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSFX_MissingPrompt(t *testing.T) {
	cmd := newSFXCmd()
	_, stderr, err := executeCommand(cmd, "-o", "output.mp3")

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

func TestSFX_MissingOutput(t *testing.T) {
	cmd := newSFXCmd()
	_, stderr, err := executeCommand(cmd, "explosion sound")

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

func TestSFX_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"too short", "0.1"},
		{"too long", "50"},
		{"negative", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSFXCmd()
			_, stderr, err := executeCommand(cmd, "explosion", "-o", "out.mp3", "-d", tt.duration)

			if err == nil {
				t.Fatal("expected error for invalid duration")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_duration" {
				t.Errorf("expected error code 'invalid_duration', got: %s", errorObj["code"])
			}
		})
	}
}

func TestSFX_InvalidInfluence(t *testing.T) {
	tests := []struct {
		name      string
		influence string
	}{
		{"negative", "-0.1"},
		{"too high", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSFXCmd()
			_, stderr, err := executeCommand(cmd, "explosion", "-o", "out.mp3", "--influence", tt.influence)

			if err == nil {
				t.Fatal("expected error for invalid influence")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_influence" {
				t.Errorf("expected error code 'invalid_influence', got: %s", errorObj["code"])
			}
		})
	}
}

func TestSFX_InvalidFormat(t *testing.T) {
	cmd := newSFXCmd()
	_, stderr, err := executeCommand(cmd, "explosion", "-o", "out.mp3", "-f", "invalid_format")

	if err == nil {
		t.Fatal("expected error for invalid format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_format" {
		t.Errorf("expected error code 'invalid_format', got: %s", errorObj["code"])
	}
}

func TestSFX_MissingAPIKey(t *testing.T) {
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newSFXCmd()
	_, stderr, err := executeCommand(cmd, "explosion", "-o", "out.mp3")

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

func TestSFX_ValidFlags(t *testing.T) {
	cmd := newSFXCmd()

	if cmd.Flag("output") == nil {
		t.Error("expected --output flag")
	}
	if cmd.Flag("prompt-file") == nil {
		t.Error("expected --prompt-file flag")
	}
	if cmd.Flag("duration") == nil {
		t.Error("expected --duration flag")
	}
	if cmd.Flag("loop") == nil {
		t.Error("expected --loop flag")
	}
	if cmd.Flag("influence") == nil {
		t.Error("expected --influence flag")
	}
	if cmd.Flag("format") == nil {
		t.Error("expected --format flag")
	}
}

func TestSFX_DefaultValues(t *testing.T) {
	cmd := newSFXCmd()

	if cmd.Flag("duration").DefValue != "0" {
		t.Errorf("expected default duration '0', got: %s", cmd.Flag("duration").DefValue)
	}
	if cmd.Flag("loop").DefValue != "false" {
		t.Errorf("expected default loop 'false', got: %s", cmd.Flag("loop").DefValue)
	}
	if cmd.Flag("influence").DefValue != "0.3" {
		t.Errorf("expected default influence '0.3', got: %s", cmd.Flag("influence").DefValue)
	}
	if cmd.Flag("format").DefValue != "mp3_44100_128" {
		t.Errorf("expected default format 'mp3_44100_128', got: %s", cmd.Flag("format").DefValue)
	}
}

func TestSFX_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "sfx_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("explosion in the distance")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newSFXCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", tmpFile.Name(), "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (file read success), got: %s", errorObj["code"])
	}
}

func TestSFX_FromStdin(t *testing.T) {
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newSFXCmd()
	cmd.SetIn(strings.NewReader("thunder in the distance"))

	_, stderr, err := executeCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (stdin read success), got: %s", errorObj["code"])
	}
}

func TestSFX_ValidDurationZeroAllowed(t *testing.T) {
	// Duration 0 should be allowed (means auto)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newSFXCmd()
	_, stderr, err := executeCommand(cmd, "explosion", "-o", "out.mp3", "-d", "0")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, not duration validation
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}
