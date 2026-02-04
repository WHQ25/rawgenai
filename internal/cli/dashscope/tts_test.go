package dashscope

import (
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== Required Field Validation =====

func TestTTS_MissingText(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "-o", "output.wav")

	if err == nil {
		t.Fatal("expected error for missing text")
	}
	expectErrorCode(t, stderr, "missing_text")
}

func TestTTS_MissingOutput(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello world")

	if err == nil {
		t.Fatal("expected error for missing output")
	}
	expectErrorCode(t, stderr, "missing_output")
}

func TestTTS_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.wav")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

// ===== Format Validation =====

func TestTTS_UnsupportedFormat(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.xyz")

	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	expectErrorCode(t, stderr, "unsupported_format")
}

func TestTTS_HTTPModelOnlyWAV(t *testing.T) {
	tests := []struct {
		name string
		ext  string
	}{
		{"mp3", "out.mp3"},
		{"pcm", "out.pcm"},
		{"opus", "out.opus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", tt.ext, "-m", "qwen3-tts-flash")

			if err == nil {
				t.Fatal("expected error for non-WAV format with HTTP model")
			}
			expectErrorCode(t, stderr, "unsupported_format")
		})
	}
}

func TestTTS_RealtimeModelFormats(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tests := []struct {
		name string
		ext  string
	}{
		{"mp3", "out.mp3"},
		{"pcm", "out.pcm"},
		{"opus", "out.opus"},
		{"wav", "out.wav"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", tt.ext, "-m", "qwen3-tts-flash-realtime")

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}
			// Should pass format validation and reach API key check
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

// ===== Invalid Parameter Values =====

func TestTTS_InvalidModel(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.wav", "-m", "invalid-model")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}
	expectErrorCode(t, stderr, "invalid_model")
}

func TestTTS_InvalidLanguage(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.wav", "-l", "Klingon")

	if err == nil {
		t.Fatal("expected error for invalid language")
	}
	expectErrorCode(t, stderr, "invalid_language")
}

func TestTTS_InvalidSampleRate(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.mp3", "-m", "qwen3-tts-flash-realtime", "--sample-rate", "44100")

	if err == nil {
		t.Fatal("expected error for invalid sample rate")
	}
	expectErrorCode(t, stderr, "invalid_sample_rate")
}

// ===== Compatibility Checks =====

func TestTTS_IncompatibleInstructions(t *testing.T) {
	tests := []struct {
		name  string
		model string
	}{
		{"http flash", "qwen3-tts-flash"},
		{"realtime flash", "qwen3-tts-flash-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			ext := "out.wav"
			if isRealtimeModel(tt.model) {
				ext = "out.mp3"
			}
			_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", ext, "-m", tt.model, "--instructions", "speak fast")

			if err == nil {
				t.Fatal("expected error for instructions with non-instruct model")
			}
			expectErrorCode(t, stderr, "incompatible_instructions")
		})
	}
}

func TestTTS_InstructionsWithInstructModel(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.mp3",
		"-m", "qwen3-tts-instruct-flash-realtime",
		"--instructions", "speak fast")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	// Should pass validation and reach API key check
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestTTS_IncompatibleSampleRate(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.wav", "-m", "qwen3-tts-flash", "--sample-rate", "48000")

	if err == nil {
		t.Fatal("expected error for sample rate with HTTP model")
	}
	expectErrorCode(t, stderr, "incompatible_sample_rate")
}

func TestTTS_TextTooLong(t *testing.T) {
	// 601 characters
	longText := strings.Repeat("你", 301) // 301 Chinese chars = 602 characters
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, longText, "-o", "out.wav")

	if err == nil {
		t.Fatal("expected error for text too long")
	}
	expectErrorCode(t, stderr, "text_too_long")
}

func TestTTS_TextTooLongRealtime(t *testing.T) {
	// Realtime model should not check text length (streaming handles it)
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	longText := strings.Repeat("你", 301)
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, longText, "-o", "out.mp3", "-m", "qwen3-tts-flash-realtime")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

// ===== Input Sources =====

func TestTTS_FromFile(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "tts_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("Hello from file")
	tmpFile.Close()

	cmd := newTTSCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "-f", tmpFile.Name(), "-o", "out.wav")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestTTS_FileNotFound(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "-f", "/nonexistent/file.txt", "-o", "out.wav")

	if err == nil {
		t.Fatal("expected error for file not found")
	}
	expectErrorCode(t, stderr, "missing_text")
}

func TestTTS_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newTTSCmd()
	cmd.SetIn(strings.NewReader("Hello from stdin"))

	_, stderr, err := executeVideoCommand(cmd, "-o", "out.wav")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

// ===== --speak flag =====

func TestTTS_SpeakWithoutOutput(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeVideoCommand(cmd, "Hello", "--speak")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	// Should not trigger missing_output
	expectErrorCode(t, stderr, "missing_api_key")
}

// ===== Valid Languages =====

func TestTTS_ValidLanguages(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	languages := []string{"Auto", "Chinese", "English", "Japanese", "Korean", "French", "German", "Russian", "Italian", "Spanish", "Portuguese"}
	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.wav", "-l", lang)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

// ===== Valid Sample Rates =====

func TestTTS_ValidSampleRates(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	sampleRates := []string{"24000", "48000"}
	for _, sr := range sampleRates {
		t.Run(sr, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeVideoCommand(cmd, "Hello", "-o", "out.mp3", "-m", "qwen3-tts-flash-realtime", "--sample-rate", sr)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

// ===== Flag Registration =====

func TestTTS_AllFlags(t *testing.T) {
	cmd := newTTSCmd()

	flags := []string{
		"output", "file", "voice", "model",
		"language", "instructions", "sample-rate", "speak",
	}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestTTS_ShortFlags(t *testing.T) {
	cmd := newTTSCmd()

	shortFlags := map[string]string{
		"o": "output",
		"f": "file",
		"m": "model",
		"l": "language",
	}

	for short, long := range shortFlags {
		flag := cmd.Flag(long)
		if flag == nil {
			t.Errorf("flag --%s not found", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected short flag '-%s' for '--%s', got '-%s'", short, long, flag.Shorthand)
		}
	}
}

func TestTTS_DefaultValues(t *testing.T) {
	cmd := newTTSCmd()

	defaults := map[string]string{
		"voice":       "Cherry",
		"model":       "qwen3-tts-flash",
		"language":    "Auto",
		"sample-rate": "24000",
		"speak":       "false",
	}

	for flag, expected := range defaults {
		f := cmd.Flag(flag)
		if f == nil {
			t.Errorf("flag --%s not found", flag)
			continue
		}
		if f.DefValue != expected {
			t.Errorf("expected default %s '%s', got: %s", flag, expected, f.DefValue)
		}
	}
}
