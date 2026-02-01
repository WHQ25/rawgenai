package google

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

func TestTTS_MissingText(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"-o", "output.wav"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "missing_text" {
		t.Errorf("expected error code 'missing_text', got '%s'", resp.Error.Code)
	}
}

func TestTTS_MissingOutput(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello world"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "missing_output" {
		t.Errorf("expected error code 'missing_output', got '%s'", resp.Error.Code)
	}
}

func TestTTS_UnsupportedFormat(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello", "-o", "output.mp3"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "unsupported_format" {
		t.Errorf("expected error code 'unsupported_format', got '%s'", resp.Error.Code)
	}
}

func TestTTS_InvalidModel(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello", "-o", "output.wav", "-m", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "invalid_model" {
		t.Errorf("expected error code 'invalid_model', got '%s'", resp.Error.Code)
	}
}

func TestTTS_InvalidVoice(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello", "-o", "output.wav", "-v", "InvalidVoice"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "invalid_voice" {
		t.Errorf("expected error code 'invalid_voice', got '%s'", resp.Error.Code)
	}
}

func TestTTS_ConflictingFlags(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello", "-o", "output.wav", "-v", "Puck", "--speakers", "Joe=Kore"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "conflicting_flags" {
		t.Errorf("expected error code 'conflicting_flags', got '%s'", resp.Error.Code)
	}
}

func TestTTS_TooManySpeakers(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello", "-o", "output.wav", "--speakers", "Joe=Kore,Jane=Puck,Bob=Leda"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "too_many_speakers" {
		t.Errorf("expected error code 'too_many_speakers', got '%s'", resp.Error.Code)
	}
}

func TestTTS_InvalidSpeakerVoice(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello", "-o", "output.wav", "--speakers", "Joe=InvalidVoice"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "invalid_speakers" {
		t.Errorf("expected error code 'invalid_speakers', got '%s'", resp.Error.Code)
	}
}

func TestTTS_InvalidSpeakerFormat(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"Hello", "-o", "output.wav", "--speakers", "JoeKore"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "invalid_speakers" {
		t.Errorf("expected error code 'invalid_speakers', got '%s'", resp.Error.Code)
	}
}

func TestTTS_FromStdin(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetIn(strings.NewReader("Hello from stdin"))
	cmd.SetArgs([]string{"-o", "output.wav"})

	// Will fail at API key check, but text should be read successfully
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error (missing API key), got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	// Should reach API key validation (text was read successfully)
	if resp.Error.Code != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got '%s'", resp.Error.Code)
	}
}

func TestParseSpeakers(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "single speaker",
			input: "Joe=Kore",
			want:  map[string]string{"Joe": "Kore"},
		},
		{
			name:  "two speakers",
			input: "Joe=Kore,Jane=Puck",
			want:  map[string]string{"Joe": "Kore", "Jane": "Puck"},
		},
		{
			name:  "with spaces",
			input: " Joe = Kore , Jane = Puck ",
			want:  map[string]string{"Joe": "Kore", "Jane": "Puck"},
		},
		{
			name:    "invalid format",
			input:   "JoeKore",
			wantErr: true,
		},
		{
			name:    "invalid voice",
			input:   "Joe=InvalidVoice",
			wantErr: true,
		},
		{
			name:    "empty speaker name",
			input:   "=Kore",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSpeakers(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("got %d speakers, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("speaker %s: got %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestPcmToWAV(t *testing.T) {
	// Simple test with known input
	pcm := []byte{0x01, 0x02, 0x03, 0x04}
	wav := pcmToWAV(pcm, 24000, 16, 1)

	// Check header
	if string(wav[0:4]) != "RIFF" {
		t.Error("expected RIFF header")
	}
	if string(wav[8:12]) != "WAVE" {
		t.Error("expected WAVE format")
	}
	if string(wav[12:16]) != "fmt " {
		t.Error("expected fmt subchunk")
	}
	if string(wav[36:40]) != "data" {
		t.Error("expected data subchunk")
	}

	// Check data
	if !bytes.Equal(wav[44:], pcm) {
		t.Error("PCM data not correctly appended")
	}

	// Total size should be 44 + len(pcm)
	if len(wav) != 44+len(pcm) {
		t.Errorf("expected length %d, got %d", 44+len(pcm), len(wav))
	}
}

func TestTTS_FlagDefaults(t *testing.T) {
	cmd := newTTSCmd()

	// Check default values
	voice, _ := cmd.Flags().GetString("voice")
	if voice != "Kore" {
		t.Errorf("expected default voice 'Kore', got '%s'", voice)
	}

	model, _ := cmd.Flags().GetString("model")
	if model != "flash" {
		t.Errorf("expected default model 'flash', got '%s'", model)
	}
}

func TestTTS_FileNotFound(t *testing.T) {
	cmd := newTTSCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--file", "nonexistent.txt", "-o", "output.wav"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "missing_text" {
		t.Errorf("expected error code 'missing_text', got '%s'", resp.Error.Code)
	}
}

func TestTTS_ValidVoices(t *testing.T) {
	voices := []string{
		"Zephyr", "Puck", "Charon", "Kore", "Fenrir",
		"Leda", "Orus", "Aoede", "Callirrhoe", "Autonoe",
		"Enceladus", "Iapetus", "Umbriel", "Algieba", "Despina",
		"Erinome", "Algenib", "Rasalgethi", "Laomedeia", "Achernar",
		"Alnilam", "Schedar", "Gacrux", "Pulcherrima", "Achird",
		"Zubenelgenubi", "Vindemiatrix", "Sadachbia", "Sadaltager", "Sulafat",
	}

	for _, voice := range voices {
		if !validVoices[voice] {
			t.Errorf("expected voice '%s' to be valid", voice)
		}
	}

	// Verify count
	if len(validVoices) != 30 {
		t.Errorf("expected 30 valid voices, got %d", len(validVoices))
	}
}

func TestTTS_ValidModels(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
	}{
		{"flash", "gemini-2.5-flash-preview-tts"},
		{"pro", "gemini-2.5-pro-preview-tts"},
	}

	for _, tt := range tests {
		if got := ttsModelIDs[tt.name]; got != tt.modelID {
			t.Errorf("model %s: expected %s, got %s", tt.name, tt.modelID, got)
		}
	}
}
