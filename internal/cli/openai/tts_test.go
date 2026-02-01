package openai

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestTTS_MissingText(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "-o", "output.mp3")

	if err == nil {
		t.Fatal("expected error for missing text")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_text" {
		t.Errorf("expected error code 'missing_text', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingOutput(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello world")

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

func TestTTS_InvalidSpeed(t *testing.T) {
	tests := []struct {
		name  string
		speed string
	}{
		{"too slow", "0.1"},
		{"too fast", "5.0"},
		{"negative", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--speed", tt.speed)

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

func TestTTS_UnsupportedFormat(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.xyz")

	if err == nil {
		t.Fatal("expected error for unsupported format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "unsupported_format" {
		t.Errorf("expected error code 'unsupported_format', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingAPIKey(t *testing.T) {
	// Temporarily unset API key
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3")

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

func TestTTS_ValidFlags(t *testing.T) {
	cmd := newTTSCmd()

	// Just check that flags are registered correctly
	if cmd.Flag("output") == nil {
		t.Error("expected --output flag")
	}
	if cmd.Flag("file") == nil {
		t.Error("expected --file flag")
	}
	if cmd.Flag("voice") == nil {
		t.Error("expected --voice flag")
	}
	if cmd.Flag("model") == nil {
		t.Error("expected --model flag")
	}
	if cmd.Flag("instructions") == nil {
		t.Error("expected --instructions flag")
	}
	if cmd.Flag("speed") == nil {
		t.Error("expected --speed flag")
	}
}

func TestTTS_DefaultValues(t *testing.T) {
	cmd := newTTSCmd()

	if cmd.Flag("voice").DefValue != "alloy" {
		t.Errorf("expected default voice 'alloy', got: %s", cmd.Flag("voice").DefValue)
	}
	if cmd.Flag("model").DefValue != "gpt-4o-mini-tts" {
		t.Errorf("expected default model 'gpt-4o-mini-tts', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("speed").DefValue != "1" {
		t.Errorf("expected default speed '1', got: %s", cmd.Flag("speed").DefValue)
	}
}

func TestTTS_FromFile(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "tts_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("Hello from file")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	t.Setenv("OPENAI_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name(), "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, meaning file was read successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (file read success), got: %s", errorObj["code"])
	}
}

func TestTTS_FromFileNotFound(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "--file", "/nonexistent/file.txt", "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for file not found")
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

func TestTTS_FromStdin(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newTTSCmd()
	cmd.SetIn(strings.NewReader("Hello from stdin"))

	_, stderr, err := executeCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, meaning stdin was read successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (stdin read success), got: %s", errorObj["code"])
	}
}

func TestTTS_InstructionsNotSupportedByModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
	}{
		{"tts-1", "tts-1"},
		{"tts-1-hd", "tts-1-hd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--model", tt.model, "--instructions", "Speak slowly")

			if err == nil {
				t.Fatal("expected error for instructions with unsupported model")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_parameter" {
				t.Errorf("expected error code 'invalid_parameter', got: %s", errorObj["code"])
			}
		})
	}
}
