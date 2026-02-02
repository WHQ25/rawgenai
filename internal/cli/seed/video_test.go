package seed

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

// ===== Create Command Tests =====

func TestVideoCreate_MissingPrompt(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create")

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

func TestVideoCreate_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat playing piano")

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

func TestVideoCreate_InvalidRatio(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--ratio", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid ratio")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_ratio" {
		t.Errorf("expected error code 'invalid_ratio', got: %s", errorObj["code"])
	}
}

func TestVideoCreate_InvalidResolution(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--resolution", "4k")

	if err == nil {
		t.Fatal("expected error for invalid resolution")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_resolution" {
		t.Errorf("expected error code 'invalid_resolution', got: %s", errorObj["code"])
	}
}

func TestVideoCreate_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"too_short", "3"},
		{"too_long", "13"},
		{"zero", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--duration", tt.duration)

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

func TestVideoCreate_FirstFrameNotFound(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--first-frame", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for first frame not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "first_frame_not_found" {
		t.Errorf("expected error code 'first_frame_not_found', got: %s", errorObj["code"])
	}
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

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "last_frame_requires_first" {
		t.Errorf("expected error code 'last_frame_requires_first', got: %s", errorObj["code"])
	}
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

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "last_frame_not_found" {
		t.Errorf("expected error code 'last_frame_not_found', got: %s", errorObj["code"])
	}
}

func TestVideoCreate_ValidRatios(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	ratios := []string{"16:9", "9:16", "4:3", "3:4", "1:1", "21:9"}
	for _, ratio := range ratios {
		t.Run(ratio, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--ratio", ratio)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected ratio '%s' to be valid, got error: %s", ratio, errorObj["code"])
			}
		})
	}
}

func TestVideoCreate_ValidResolutions(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	resolutions := []string{"480p", "720p", "1080p"}
	for _, resolution := range resolutions {
		t.Run(resolution, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--resolution", resolution)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected resolution '%s' to be valid, got error: %s", resolution, errorObj["code"])
			}
		})
	}
}

func TestVideoCreate_ValidDurations(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	durations := []string{"4", "5", "8", "10", "12"}
	for _, duration := range durations {
		t.Run(duration+"s", func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "create", "A cat", "--duration", duration)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected duration '%s' to be valid, got error: %s", duration, errorObj["code"])
			}
		})
	}
}

func TestVideoCreate_WithFirstFrame(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newVideoCmd()
	_, stderr, cmdErr := executeVideoCommand(cmd, "create", "A cat walking", "--first-frame", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected first frame to be valid, got error: %s", errorObj["code"])
	}
}

func TestVideoCreate_WithFirstAndLastFrame(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

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
	_, stderr, cmdErr := executeVideoCommand(cmd, "create", "Camera pan", "--first-frame", firstFrame.Name(), "--last-frame", lastFrame.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected first and last frame to be valid, got error: %s", errorObj["code"])
	}
}

func TestVideoCreate_FromFile(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

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

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected prompt from file to be valid, got error: %s", errorObj["code"])
	}
}

func TestVideoCreate_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	cmd := newVideoCmd()
	cmd.SetIn(strings.NewReader("A beautiful sunset"))

	_, stderr, err := executeVideoCommand(cmd, "create")

	if err == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected prompt from stdin to be valid, got error: %s", errorObj["code"])
	}
}

func TestVideoCreate_DefaultValues(t *testing.T) {
	cmd := newVideoCreateCmd()

	if cmd.Flag("ratio").DefValue != "16:9" {
		t.Errorf("expected default ratio '16:9', got: %s", cmd.Flag("ratio").DefValue)
	}
	if cmd.Flag("resolution").DefValue != "1080p" {
		t.Errorf("expected default resolution '1080p', got: %s", cmd.Flag("resolution").DefValue)
	}
	if cmd.Flag("duration").DefValue != "5" {
		t.Errorf("expected default duration '5', got: %s", cmd.Flag("duration").DefValue)
	}
	if cmd.Flag("audio").DefValue != "false" {
		t.Errorf("expected default audio 'false', got: %s", cmd.Flag("audio").DefValue)
	}
	if cmd.Flag("watermark").DefValue != "false" {
		t.Errorf("expected default watermark 'false', got: %s", cmd.Flag("watermark").DefValue)
	}
	if cmd.Flag("return-last-frame").DefValue != "false" {
		t.Errorf("expected default return-last-frame 'false', got: %s", cmd.Flag("return-last-frame").DefValue)
	}
}

func TestVideoCreate_ShortFlags(t *testing.T) {
	cmd := newVideoCreateCmd()

	shortFlags := map[string]string{
		"r": "ratio",
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

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

func TestVideoStatus_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "status", "cgt-2025xxxx")

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

// ===== Download Command Tests =====

func TestVideoDownload_MissingTaskID(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "-o", "output.mp4")

	if err == nil {
		t.Fatal("expected error for missing task ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

func TestVideoDownload_MissingOutput(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "cgt-2025xxxx")

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

func TestVideoDownload_InvalidFormat(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "cgt-2025xxxx", "-o", "output.avi")

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

func TestVideoDownload_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "download", "cgt-2025xxxx", "-o", "output.mp4")

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

// ===== List Command Tests =====

func TestVideoList_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "list")

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

func TestVideoList_InvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too_large", "101"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "list", "--limit", tt.limit)

			if err == nil {
				t.Fatal("expected error for invalid limit")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_limit" {
				t.Errorf("expected error code 'invalid_limit', got: %s", errorObj["code"])
			}
		})
	}
}

func TestVideoList_InvalidStatus(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "list", "--status", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid status")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_status" {
		t.Errorf("expected error code 'invalid_status', got: %s", errorObj["code"])
	}
}

func TestVideoList_ValidStatuses(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	statuses := []string{"queued", "running", "succeeded", "failed"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			cmd := newVideoCmd()
			_, stderr, err := executeVideoCommand(cmd, "list", "--status", status)

			if err == nil {
				t.Fatal("expected error (missing api key)")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected status '%s' to be valid, got error: %s", status, errorObj["code"])
			}
		})
	}
}

func TestVideoList_DefaultValues(t *testing.T) {
	cmd := newVideoListCmd()

	if cmd.Flag("limit").DefValue != "20" {
		t.Errorf("expected default limit '20', got: %s", cmd.Flag("limit").DefValue)
	}
}

// ===== Delete Command Tests =====

func TestVideoDelete_MissingTaskID(t *testing.T) {
	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "delete")

	if err == nil {
		t.Fatal("expected error for missing task ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

func TestVideoDelete_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ARK_API_KEY", "")

	cmd := newVideoCmd()
	_, stderr, err := executeVideoCommand(cmd, "delete", "cgt-2025xxxx")

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

// ===== Flags Tests =====

func TestVideoCreate_AllFlags(t *testing.T) {
	cmd := newVideoCreateCmd()

	flags := []string{"prompt-file", "first-frame", "last-frame", "ratio", "resolution", "duration", "audio", "seed", "watermark", "return-last-frame"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestVideoDownload_AllFlags(t *testing.T) {
	cmd := newVideoDownloadCmd()

	flags := []string{"output", "last-frame"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestVideoList_AllFlags(t *testing.T) {
	cmd := newVideoListCmd()

	flags := []string{"limit", "status"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}
