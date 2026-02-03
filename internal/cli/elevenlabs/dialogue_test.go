package elevenlabs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeDialogueCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()

	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestDialogue_MissingInput(t *testing.T) {
	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(""))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for missing input")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_input" {
		t.Errorf("expected error code 'missing_input', got '%s'", errorObj["code"])
	}
}

func TestDialogue_MissingOutput(t *testing.T) {
	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(`[{"text":"Hello","voice_id":"Rachel"}]`))
	_, stderr, err := executeDialogueCommand(cmd)

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

func TestDialogue_InvalidInput(t *testing.T) {
	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader("not valid json"))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for invalid JSON input")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_input" {
		t.Errorf("expected error code 'invalid_input', got '%s'", errorObj["code"])
	}
}

func TestDialogue_EmptyInput(t *testing.T) {
	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader("[]"))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for empty input array")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "empty_input" {
		t.Errorf("expected error code 'empty_input', got '%s'", errorObj["code"])
	}
}

func TestDialogue_EmptyText(t *testing.T) {
	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(`[{"text":"","voice_id":"Rachel"}]`))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for empty text")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "empty_text" {
		t.Errorf("expected error code 'empty_text', got '%s'", errorObj["code"])
	}
}

func TestDialogue_TooManyVoices(t *testing.T) {
	// Create input with 11 unique voices
	inputs := make([]map[string]string, 11)
	for i := 0; i < 11; i++ {
		inputs[i] = map[string]string{
			"text":     "Hello",
			"voice_id": string(rune('A'+i)) + "voice",
		}
	}
	inputJSON, _ := json.Marshal(inputs)

	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(string(inputJSON)))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for too many voices")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "too_many_voices" {
		t.Errorf("expected error code 'too_many_voices', got '%s'", errorObj["code"])
	}
}

func TestDialogue_InvalidStability(t *testing.T) {
	tests := []struct {
		name      string
		stability string
	}{
		{"negative", "-0.5"},
		{"too_high", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newDialogueCmd()
			cmd.SetIn(strings.NewReader(`[{"text":"Hello","voice_id":"Rachel"}]`))
			_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3", "--stability", tt.stability)

			if err == nil {
				t.Fatal("expected error for invalid stability")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_stability" {
				t.Errorf("expected error code 'invalid_stability', got '%s'", errorObj["code"])
			}
		})
	}
}

func TestDialogue_InvalidTextNormalization(t *testing.T) {
	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(`[{"text":"Hello","voice_id":"Rachel"}]`))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3", "--text-normalization", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid text normalization")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_text_normalization" {
		t.Errorf("expected error code 'invalid_text_normalization', got '%s'", errorObj["code"])
	}
}

func TestDialogue_InvalidFormat(t *testing.T) {
	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(`[{"text":"Hello","voice_id":"Rachel"}]`))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3", "-f", "invalid_format")

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

func TestDialogue_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(`[{"text":"Hello","voice_id":"Rachel"}]`))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3")

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

func TestDialogue_ValidFlags(t *testing.T) {
	cmd := newDialogueCmd()

	flags := []string{"output", "input", "model", "format", "language", "stability", "text-normalization", "seed", "speak"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag '%s' not found", flag)
		}
	}

	shortFlags := map[string]string{
		"o": "output",
		"i": "input",
		"m": "model",
		"f": "format",
		"l": "language",
	}
	for short, long := range shortFlags {
		flag := cmd.Flags().Lookup(long)
		if flag == nil || flag.Shorthand != short {
			t.Errorf("short flag '-%s' not mapped to '--%s'", short, long)
		}
	}
}

func TestDialogue_DefaultValues(t *testing.T) {
	cmd := newDialogueCmd()

	tests := []struct {
		flag     string
		expected string
	}{
		{"model", "eleven_v3"},
		{"format", "mp3_44100_128"},
		{"stability", "0.5"},
		{"text-normalization", "auto"},
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

func TestDialogue_VoiceNameResolution(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newDialogueCmd()
	cmd.SetIn(strings.NewReader(`[{"text":"Hello","voice_id":"Rachel"},{"text":"Hi","voice_id":"Josh"}]`))
	_, stderr, err := executeDialogueCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing API key)")
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
