package dashscope

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeVideoCommand(cmd *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func expectErrorCode(t *testing.T, stderr string, expectedCode string) {
	t.Helper()
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
	if resp["success"] != false {
		t.Error("expected success to be false")
	}
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != expectedCode {
		t.Errorf("expected error code '%s', got: %s", expectedCode, errorObj["code"])
	}
}

// ===== Create Command Tests =====

func TestVideoCreate_MissingPrompt(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create")

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}
	expectErrorCode(t, stderr, "missing_prompt")
}

func TestVideoCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat playing piano")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_InvalidResolution(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--resolution", "4K")

	if err == nil {
		t.Fatal("expected error for invalid resolution")
	}
	expectErrorCode(t, stderr, "invalid_resolution")
}

func TestVideoCreate_InvalidRatio(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--ratio", "4:3")

	if err == nil {
		t.Fatal("expected error for invalid ratio")
	}
	expectErrorCode(t, stderr, "invalid_ratio")
}

func TestVideoCreate_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"too_short", "1"},
		{"too_long", "20"},
		{"zero", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--duration", tt.duration)

			if err == nil {
				t.Fatal("expected error for invalid duration")
			}
			expectErrorCode(t, stderr, "invalid_duration")
		})
	}
}

func TestVideoCreate_InvalidModel(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--model", "invalid-model")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}
	expectErrorCode(t, stderr, "invalid_model")
}

func TestVideoCreate_ImageNotFound(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--image", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for image not found")
	}
	expectErrorCode(t, stderr, "image_not_found")
}

func TestVideoCreate_ImageURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--image", "https://example.com/image.jpg")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_FirstFrameNotFound(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--first-frame", "/nonexistent/frame.jpg")

	if err == nil {
		t.Fatal("expected error for first frame not found")
	}
	expectErrorCode(t, stderr, "first_frame_not_found")
}

func TestVideoCreate_LastFrameRequiresFirst(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newVideoCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "create", "A cat", "--last-frame", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error for last frame without first frame")
	}
	expectErrorCode(t, stderr, "last_frame_requires_first")
}

func TestVideoCreate_LastFrameNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "first_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newVideoCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "create", "A cat", "--first-frame", tmpFile.Name(), "--last-frame", "/nonexistent/last.jpg")

	if cmdErr == nil {
		t.Fatal("expected error for last frame not found")
	}
	expectErrorCode(t, stderr, "last_frame_not_found")
}

func TestVideoCreate_ConflictingInputFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"image_and_first_frame", []string{"create", "A cat", "--image", "https://example.com/img.jpg", "--first-frame", "https://example.com/frame.jpg"}},
		{"image_and_ref", []string{"create", "A cat", "--image", "https://example.com/img.jpg", "--ref", "https://example.com/ref.jpg"}},
		{"ref_and_first_frame", []string{"create", "A cat", "--ref", "https://example.com/ref.jpg", "--first-frame", "https://example.com/frame.jpg"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, tt.args...)

			if err == nil {
				t.Fatal("expected error for conflicting input flags")
			}
			expectErrorCode(t, stderr, "conflicting_input_flags")
		})
	}
}

func TestVideoCreate_RefNotFound(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "character1 walks", "--ref", "/nonexistent/ref.jpg")

	if err == nil {
		t.Fatal("expected error for ref not found")
	}
	expectErrorCode(t, stderr, "ref_not_found")
}

func TestVideoCreate_RefURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "character1 walks", "--ref", "https://example.com/person.jpg")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_TooManyRefs(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "character1 walks",
		"--ref", "https://example.com/1.jpg",
		"--ref", "https://example.com/2.jpg",
		"--ref", "https://example.com/3.jpg",
		"--ref", "https://example.com/4.jpg",
		"--ref", "https://example.com/5.jpg",
		"--ref", "https://example.com/6.jpg",
	)

	if err == nil {
		t.Fatal("expected error for too many refs")
	}
	expectErrorCode(t, stderr, "too_many_refs")
}

func TestVideoCreate_IncompatibleAudio(t *testing.T) {
	// --audio only works with wan2.6-i2v-flash
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--audio")

	if err == nil {
		t.Fatal("expected error for incompatible audio")
	}
	expectErrorCode(t, stderr, "incompatible_audio")
}

func TestVideoCreate_IncompatibleNoAudio(t *testing.T) {
	// --no-audio only works with wan2.6-r2v-flash
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--no-audio")

	if err == nil {
		t.Fatal("expected error for incompatible no-audio")
	}
	expectErrorCode(t, stderr, "incompatible_no_audio")
}

func TestVideoCreate_IncompatibleShotType(t *testing.T) {
	// --shot-type only works with wan2.6 models
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--model", "wan2.2-t2v-plus", "--shot-type", "multi")

	if err == nil {
		t.Fatal("expected error for incompatible shot type")
	}
	expectErrorCode(t, stderr, "incompatible_shot_type")
}

func TestVideoCreate_IncompatibleAudioURL(t *testing.T) {
	// --audio-url only works with wan2.5+ t2v/i2v models, not r2v
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "character1 walks",
		"--ref", "https://example.com/person.jpg",
		"--audio-url", "https://example.com/audio.wav",
	)

	if err == nil {
		t.Fatal("expected error for incompatible audio url")
	}
	expectErrorCode(t, stderr, "incompatible_audio_url")
}

func TestVideoCreate_ValidResolutions(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	resolutions := []string{"480P", "720P", "1080P"}
	for _, res := range resolutions {
		t.Run(res, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--resolution", res)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

func TestVideoCreate_ValidRatios(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	ratios := []string{"16:9", "9:16"}
	for _, ratio := range ratios {
		t.Run(ratio, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--ratio", ratio)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

func TestVideoCreate_ValidDurations(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	durations := []string{"2", "5", "10", "15"}
	for _, duration := range durations {
		t.Run(duration+"s", func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--duration", duration)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}
			expectErrorCode(t, stderr, "missing_api_key")
		})
	}
}

func TestVideoCreate_WithImage(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "Make it move", "--image", "https://example.com/photo.jpg")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_WithRef(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "character1 walks", "--ref", "https://example.com/person.jpg")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_WithMultipleRefs(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "character1 和 character2 聊天",
		"--ref", "https://example.com/person1.jpg",
		"--ref", "https://example.com/person2.jpg",
	)

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_WithFirstFrame(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newVideoCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "create", "Camera push in", "--first-frame", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_WithFirstAndLastFrame(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	firstFrame, err := os.CreateTemp("", "first_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(firstFrame.Name())
	firstFrame.Close()

	lastFrame, err := os.CreateTemp("", "last_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lastFrame.Name())
	lastFrame.Close()

	cmd := newVideoCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "create", "Transition", "--first-frame", firstFrame.Name(), "--last-frame", lastFrame.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_FromFile(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("A beautiful sunset over the ocean")
	tmpFile.Close()

	cmd := newVideoCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "create", "--prompt-file", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	cmd.SetIn(strings.NewReader("A beautiful sunset"))

	_, stderr, err := executeVideoCommand(cmd, "create")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoCreate_DefaultValues(t *testing.T) {
	cmd := newVideoCreateCmd()

	defaults := map[string]string{
		"resolution":    "720P",
		"ratio":         "16:9",
		"duration":      "5",
		"audio":         "false",
		"no-audio":      "false",
		"prompt-extend": "true",
		"watermark":     "false",
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

func TestVideoCreate_AllFlags(t *testing.T) {
	cmd := newVideoCreateCmd()

	flags := []string{
		"image", "ref", "first-frame", "last-frame",
		"prompt-file", "model", "resolution", "ratio",
		"duration", "negative", "audio", "no-audio",
		"audio-url", "shot-type", "prompt-extend",
		"watermark", "seed",
	}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestVideoCreate_ShortFlags(t *testing.T) {
	cmd := newVideoCreateCmd()

	shortFlags := map[string]string{
		"i": "image",
		"f": "prompt-file",
		"m": "model",
		"r": "resolution",
		"d": "duration",
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

// ===== Status Command Tests =====

func TestVideoStatus_MissingTaskID(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "status")

	if err == nil {
		t.Fatal("expected error for missing task ID")
	}
	expectErrorCode(t, stderr, "missing_task_id")
}

func TestVideoStatus_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "status", "task-xxxx")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoStatus_AllFlags(t *testing.T) {
	cmd := newVideoStatusCmd()

	if cmd.Flag("verbose") == nil {
		t.Error("expected --verbose flag")
	}
}

func TestVideoStatus_ShortFlags(t *testing.T) {
	cmd := newVideoStatusCmd()

	flag := cmd.Flag("verbose")
	if flag == nil {
		t.Fatal("flag --verbose not found")
	}
	if flag.Shorthand != "v" {
		t.Errorf("expected short flag '-v' for '--verbose', got '-%s'", flag.Shorthand)
	}
}

// ===== Download Command Tests =====

func TestVideoDownload_MissingTaskID(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "-o", "output.mp4")

	if err == nil {
		t.Fatal("expected error for missing task ID")
	}
	expectErrorCode(t, stderr, "missing_task_id")
}

func TestVideoDownload_MissingOutput(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "task-xxxx")

	if err == nil {
		t.Fatal("expected error for missing output")
	}
	expectErrorCode(t, stderr, "missing_output")
}

func TestVideoDownload_InvalidFormat(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "task-xxxx", "-o", "output.avi")

	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	expectErrorCode(t, stderr, "invalid_format")
}

func TestVideoDownload_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("DASHSCOPE_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "task-xxxx", "-o", "output.mp4")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	expectErrorCode(t, stderr, "missing_api_key")
}

func TestVideoDownload_AllFlags(t *testing.T) {
	cmd := newVideoDownloadCmd()

	if cmd.Flag("output") == nil {
		t.Error("expected --output flag")
	}
}

func TestVideoDownload_ShortFlags(t *testing.T) {
	cmd := newVideoDownloadCmd()

	flag := cmd.Flag("output")
	if flag == nil {
		t.Fatal("flag --output not found")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected short flag '-o' for '--output', got '-%s'", flag.Shorthand)
	}
}
