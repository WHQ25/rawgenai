package dashscope

import (
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== STT Default Command: Required Field Validation =====

func TestSTT_MissingInput(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing input")
	}
	expectErrorCode(t, stderr, "missing_input")
}

func TestSTT_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	// Create a temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio data"))
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error for missing API key")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

// ===== STT Default Command: File Validation =====

func TestSTT_FileNotFound(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "/nonexistent/audio.wav")

	if err == nil {
		t.Fatal("expected error for file not found")
	}
	expectErrorCode(t, stderr, "file_not_found")
}

func TestSTT_FileTooLarge(t *testing.T) {
	// Create a file that appears > 10 MB via sparse file
	tmpFile, err := os.CreateTemp("", "stt_test_large_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	// Write 1 byte at 11MB offset to create sparse file
	if _, err := tmpFile.WriteAt([]byte{0}, 11*1024*1024); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error for file too large")
	}
	expectErrorCode(t, stderr, "file_too_large")
}

func TestSTT_FileTooLargeRealtimeSkipped(t *testing.T) {
	// Realtime models should NOT check file size
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_large_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteAt([]byte{0}, 11*1024*1024); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", "paraformer-realtime-v2")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}
	// Should pass file size check and reach API key check
	expectErrorCode(t, stderr, "missing_api_key")
}

// ===== STT Default Command: Invalid Parameters =====

func TestSTT_InvalidModel(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", "invalid-model")

	if cmdErr == nil {
		t.Fatal("expected error for invalid model")
	}
	expectErrorCode(t, stderr, "invalid_model")
}

// ===== STT Default Command: Compatibility Checks =====

func TestSTT_IncompatibleLanguage(t *testing.T) {
	// --language should not work with paraformer-realtime (use --language-hints)
	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"paraformer-realtime-v2", "paraformer-realtime-v2"},
		{"fun-asr-realtime", "fun-asr-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "-l", "zh")

			if cmdErr == nil {
				t.Fatal("expected error for incompatible language")
			}
			expectErrorCode(t, stderr, "incompatible_language")
		})
	}
}

func TestSTT_CompatibleLanguage(t *testing.T) {
	// --language should work with qwen3-asr-flash and qwen3-asr-flash-realtime
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"qwen3-asr-flash", "qwen3-asr-flash"},
		{"qwen3-asr-flash-realtime", "qwen3-asr-flash-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "-l", "zh")

			if cmdErr == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

func TestSTT_IncompatibleLanguageHints(t *testing.T) {
	// --language-hints should not work with qwen3-asr-flash
	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"qwen3-asr-flash", "qwen3-asr-flash"},
		{"qwen3-asr-flash-realtime", "qwen3-asr-flash-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "--language-hints", "zh,en")

			if cmdErr == nil {
				t.Fatal("expected error for incompatible language hints")
			}
			expectErrorCode(t, stderr, "incompatible_language_hints")
		})
	}
}

func TestSTT_CompatibleLanguageHints(t *testing.T) {
	// --language-hints should work with paraformer-realtime-v2 and fun-asr-realtime
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"paraformer-realtime-v2", "paraformer-realtime-v2"},
		{"fun-asr-realtime", "fun-asr-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "--language-hints", "zh,en")

			if cmdErr == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

func TestSTT_IncompatibleNoITN(t *testing.T) {
	// --no-itn only works with qwen3-asr-flash
	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", "paraformer-realtime-v2", "--no-itn")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible no-itn")
	}
	expectErrorCode(t, stderr, "incompatible_no_itn")
}

func TestSTT_CompatibleNoITN(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "--no-itn")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestSTT_IncompatibleVocabularyID(t *testing.T) {
	// --vocabulary-id not supported by qwen3-asr-flash or qwen3-asr-flash-realtime
	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"qwen3-asr-flash", "qwen3-asr-flash"},
		{"qwen3-asr-flash-realtime", "qwen3-asr-flash-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "--vocabulary-id", "vocab_xxx")

			if cmdErr == nil {
				t.Fatal("expected error for incompatible vocabulary-id")
			}
			expectErrorCode(t, stderr, "incompatible_vocabulary_id")
		})
	}
}

func TestSTT_CompatibleVocabularyID(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"paraformer-realtime-v2", "paraformer-realtime-v2"},
		{"fun-asr-realtime", "fun-asr-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "--vocabulary-id", "vocab_xxx")

			if cmdErr == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

func TestSTT_IncompatibleDisfluencyRemoval(t *testing.T) {
	// --disfluency-removal not supported by qwen3-asr-flash or qwen3-asr-flash-realtime
	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"qwen3-asr-flash", "qwen3-asr-flash"},
		{"qwen3-asr-flash-realtime", "qwen3-asr-flash-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "--disfluency-removal")

			if cmdErr == nil {
				t.Fatal("expected error for incompatible disfluency-removal")
			}
			expectErrorCode(t, stderr, "incompatible_disfluency_removal")
		})
	}
}

func TestSTT_CompatibleDisfluencyRemoval(t *testing.T) {
	// --disfluency-removal works with paraformer-realtime and fun-asr-realtime
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio"))
	tmpFile.Close()

	tests := []struct {
		name  string
		model string
	}{
		{"paraformer-realtime-v2", "paraformer-realtime-v2"},
		{"fun-asr-realtime", "fun-asr-realtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeVideoCommand(cmd, tmpFile.Name(), "-m", tt.model, "--disfluency-removal")

			if cmdErr == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

// ===== STT Default Command: Input Sources =====

func TestSTT_FromFileFlag(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte("fake audio data"))
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "-f", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestSTT_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newSTTCmd()
	cmd.SetIn(strings.NewReader("fake audio data from stdin"))

	_, stderr, err := executeVideoCommand(cmd)

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestSTT_FileFromFlagNotFound(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "-f", "/nonexistent/audio.wav")

	if err == nil {
		t.Fatal("expected error for file not found")
	}
	expectErrorCode(t, stderr, "file_not_found")
}

// ===== STT Default Command: Flag Registration =====

func TestSTT_AllFlags(t *testing.T) {
	cmd := newSTTCmd()

	flags := []string{
		"file", "model", "language", "no-itn", "verbose", "output",
		"vocabulary-id", "disfluency-removal", "language-hints", "sample-rate",
	}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestSTT_ShortFlags(t *testing.T) {
	cmd := newSTTCmd()

	shortFlags := map[string]string{
		"f": "file",
		"m": "model",
		"l": "language",
		"v": "verbose",
		"o": "output",
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

func TestSTT_DefaultValues(t *testing.T) {
	cmd := newSTTCmd()

	defaults := map[string]string{
		"model":              "qwen3-asr-flash",
		"no-itn":            "false",
		"verbose":           "false",
		"disfluency-removal": "false",
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

// ===== STT Create Command: Required Field Validation =====

func TestSTTCreate_MissingURL(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create")

	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	expectErrorCode(t, stderr, "missing_url")
}

func TestSTTCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

// ===== STT Create Command: Invalid Parameters =====

func TestSTTCreate_InvalidModel(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav", "-m", "invalid-model")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}
	expectErrorCode(t, stderr, "invalid_model")
}

func TestSTTCreate_InvalidSpeakers(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav", "--diarize", "--speakers", "1")

	if err == nil {
		t.Fatal("expected error for invalid speakers")
	}
	expectErrorCode(t, stderr, "invalid_speakers")
}

func TestSTTCreate_InvalidSpeakersHigh(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav", "--diarize", "--speakers", "101")

	if err == nil {
		t.Fatal("expected error for invalid speakers")
	}
	expectErrorCode(t, stderr, "invalid_speakers")
}

// ===== STT Create Command: Compatibility Checks =====

func TestSTTCreate_IncompatibleLanguageHints(t *testing.T) {
	// --language-hints only works with paraformer-v2
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "fun-asr", "--language-hints", "zh,en")

	if err == nil {
		t.Fatal("expected error for incompatible language-hints")
	}
	expectErrorCode(t, stderr, "incompatible_language_hints")
}

func TestSTTCreate_CompatibleLanguageHints(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "paraformer-v2", "--language-hints", "zh,en")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestSTTCreate_IncompatibleVocabularyID(t *testing.T) {
	// --vocabulary-id not supported by qwen3-asr-flash-filetrans
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "qwen3-asr-flash-filetrans", "--vocabulary-id", "vocab_xxx")

	if err == nil {
		t.Fatal("expected error for incompatible vocabulary-id")
	}
	expectErrorCode(t, stderr, "incompatible_vocabulary_id")
}

func TestSTTCreate_IncompatibleDiarize(t *testing.T) {
	// --diarize not supported by qwen3-asr-flash-filetrans
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "qwen3-asr-flash-filetrans", "--diarize")

	if err == nil {
		t.Fatal("expected error for incompatible diarize")
	}
	expectErrorCode(t, stderr, "incompatible_diarize")
}

func TestSTTCreate_IncompatibleITN(t *testing.T) {
	// --itn only works with qwen3-asr-flash-filetrans
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "paraformer-v2", "--itn")

	if err == nil {
		t.Fatal("expected error for incompatible itn")
	}
	expectErrorCode(t, stderr, "incompatible_itn")
}

func TestSTTCreate_CompatibleITN(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "qwen3-asr-flash-filetrans", "--itn")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestSTTCreate_IncompatibleWords(t *testing.T) {
	// --words only works with qwen3-asr-flash-filetrans
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "paraformer-v2", "--words")

	if err == nil {
		t.Fatal("expected error for incompatible words")
	}
	expectErrorCode(t, stderr, "incompatible_words")
}

func TestSTTCreate_SpeakersRequiresDiarize(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav", "--speakers", "3")

	if err == nil {
		t.Fatal("expected error for speakers without diarize")
	}
	expectErrorCode(t, stderr, "speakers_requires_diarize")
}

func TestSTTCreate_IncompatibleDisfluencyRemoval(t *testing.T) {
	// --disfluency-removal not supported by qwen3-asr-flash-filetrans
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
		"-m", "qwen3-asr-flash-filetrans", "--disfluency-removal")

	if err == nil {
		t.Fatal("expected error for incompatible disfluency-removal")
	}
	expectErrorCode(t, stderr, "incompatible_disfluency_removal")
}

func TestSTTCreate_CompatibleDisfluencyRemoval(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tests := []struct {
		name  string
		model string
	}{
		{"paraformer-v2", "paraformer-v2"},
		{"fun-asr", "fun-asr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "https://example.com/audio.wav",
				"-m", tt.model, "--disfluency-removal")

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

// ===== STT Create Command: Flag Registration =====

func TestSTTCreate_AllFlags(t *testing.T) {
	cmd := newSTTCmd()
	// Find the create subcommand
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create subcommand not found: %v", err)
	}

	flags := []string{
		"model", "language-hints", "vocabulary-id", "disfluency-removal",
		"diarize", "speakers", "channel", "itn", "words",
	}
	for _, flag := range flags {
		if createCmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag on create subcommand", flag)
		}
	}
}

func TestSTTCreate_DefaultValues(t *testing.T) {
	cmd := newSTTCmd()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create subcommand not found: %v", err)
	}

	defaults := map[string]string{
		"model":               "paraformer-v2",
		"disfluency-removal":  "false",
		"diarize":             "false",
		"itn":                 "false",
		"words":               "false",
	}

	for flag, expected := range defaults {
		f := createCmd.Flag(flag)
		if f == nil {
			t.Errorf("flag --%s not found", flag)
			continue
		}
		if f.DefValue != expected {
			t.Errorf("expected default %s '%s', got: %s", flag, expected, f.DefValue)
		}
	}
}

// ===== STT Status Command =====

func TestSTTStatus_MissingTaskID(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "status")

	if err == nil {
		t.Fatal("expected error for missing task ID")
	}
	expectErrorCode(t, stderr, "missing_task_id")
}

func TestSTTStatus_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newSTTCmd()
	_, stderr, err := executeVideoCommand(cmd, "status", "task-123")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestSTTStatus_AllFlags(t *testing.T) {
	cmd := newSTTCmd()
	statusCmd, _, err := cmd.Find([]string{"status"})
	if err != nil {
		t.Fatalf("status subcommand not found: %v", err)
	}

	flags := []string{"verbose", "output"}
	for _, flag := range flags {
		if statusCmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag on status subcommand", flag)
		}
	}
}
