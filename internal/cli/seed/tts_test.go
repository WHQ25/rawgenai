package seed

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

func TestTTS_InvalidFormat(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--format", "wav")

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

func TestTTS_InvalidSpeed(t *testing.T) {
	tests := []struct {
		name  string
		speed string
	}{
		{"too slow", "-51"},
		{"too fast", "101"},
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

func TestTTS_InvalidVolume(t *testing.T) {
	tests := []struct {
		name   string
		volume string
	}{
		{"too low", "-51"},
		{"too high", "101"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--volume", tt.volume)

			if err == nil {
				t.Fatal("expected error for invalid volume")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_volume" {
				t.Errorf("expected error code 'invalid_volume', got: %s", errorObj["code"])
			}
		})
	}
}

func TestTTS_MissingCredentials(t *testing.T) {
	t.Setenv("SEED_APP_ID", "")
	t.Setenv("SEED_ACCESS_TOKEN", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for missing credentials")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_credentials" {
		t.Errorf("expected error code 'missing_credentials', got: %s", errorObj["code"])
	}
}

func TestTTS_ValidFlags(t *testing.T) {
	cmd := newTTSCmd()

	flags := []string{"output", "prompt-file", "voice", "format", "sample-rate", "speed", "volume", "speak"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestTTS_DefaultValues(t *testing.T) {
	cmd := newTTSCmd()

	if cmd.Flag("voice").DefValue != "zh_female_vv_uranus_bigtts" {
		t.Errorf("expected default voice 'zh_female_vv_uranus_bigtts', got: %s", cmd.Flag("voice").DefValue)
	}
	if cmd.Flag("format").DefValue != "mp3" {
		t.Errorf("expected default format 'mp3', got: %s", cmd.Flag("format").DefValue)
	}
	if cmd.Flag("sample-rate").DefValue != "24000" {
		t.Errorf("expected default sample-rate '24000', got: %s", cmd.Flag("sample-rate").DefValue)
	}
	if cmd.Flag("speed").DefValue != "0" {
		t.Errorf("expected default speed '0', got: %s", cmd.Flag("speed").DefValue)
	}
	if cmd.Flag("volume").DefValue != "0" {
		t.Errorf("expected default volume '0', got: %s", cmd.Flag("volume").DefValue)
	}
}

func TestTTS_SpeakWithoutOutput(t *testing.T) {
	t.Setenv("SEED_APP_ID", "")
	t.Setenv("SEED_ACCESS_TOKEN", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "--speak")

	if err == nil {
		t.Fatal("expected error (missing credentials), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach credentials check, not missing_output
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_credentials" {
		t.Errorf("expected error code 'missing_credentials', got: %s", errorObj["code"])
	}
}

func TestTTS_SpeakOnlyMP3(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "--speak", "--format", "pcm")

	if err == nil {
		t.Fatal("expected error for non-mp3 format with --speak")
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

func TestTTS_FromFile(t *testing.T) {
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

	t.Setenv("SEED_APP_ID", "")
	t.Setenv("SEED_ACCESS_TOKEN", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", tmpFile.Name(), "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing credentials), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach credentials check, meaning file was read successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_credentials" {
		t.Errorf("expected error code 'missing_credentials' (file read success), got: %s", errorObj["code"])
	}
}

func TestTTS_FromFileNotFound(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", "/nonexistent/file.txt", "-o", "out.mp3")

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
	t.Setenv("SEED_APP_ID", "")
	t.Setenv("SEED_ACCESS_TOKEN", "")

	cmd := newTTSCmd()
	cmd.SetIn(strings.NewReader("Hello from stdin"))

	_, stderr, err := executeCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing credentials), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach credentials check, meaning stdin was read successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_credentials" {
		t.Errorf("expected error code 'missing_credentials' (stdin read success), got: %s", errorObj["code"])
	}
}
