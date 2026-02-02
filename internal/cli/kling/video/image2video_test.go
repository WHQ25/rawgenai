package video

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== Image2Video (create-from-image) Command Tests =====

func TestImage2Video_MissingFirstFrame(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-image", "A cat")

	if err == nil {
		t.Fatal("expected error for missing first frame")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_image" {
		t.Errorf("expected error code 'missing_image', got: %s", errorObj["code"])
	}
}

func TestImage2Video_FirstFrameNotFound(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "create-from-image", "-i", "/nonexistent/image.jpg")

	if err == nil {
		t.Fatal("expected error for first frame not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "image_not_found" {
		t.Errorf("expected error code 'image_not_found', got: %s", errorObj["code"])
	}
}

func TestImage2Video_InvalidCameraControlJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", tmpFile.Name(),
		"--model", "kling-v1-5",
		"--mode", "pro",
		"--duration", "5",
		"--camera-control", "not-valid-json")

	if cmdErr == nil {
		t.Fatal("expected error for invalid camera control JSON")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_camera_control" {
		t.Errorf("expected error code 'invalid_camera_control', got: %s", errorObj["code"])
	}
}

func TestImage2Video_InvalidDynamicMaskJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", tmpFile.Name(),
		"--dynamic-mask", "not-valid-json")

	if cmdErr == nil {
		t.Fatal("expected error for invalid dynamic mask JSON")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_dynamic_mask" {
		t.Errorf("expected error code 'invalid_dynamic_mask', got: %s", errorObj["code"])
	}
}

func TestImage2Video_StaticMaskNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", tmpFile.Name(),
		"--static-mask", "/nonexistent/mask.png")

	if cmdErr == nil {
		t.Fatal("expected error for static mask not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "mask_not_found" {
		t.Errorf("expected error code 'mask_not_found', got: %s", errorObj["code"])
	}
}

func TestImage2Video_ValidCameraControl(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", tmpFile.Name(),
		"--model", "kling-v1-5",
		"--mode", "pro",
		"--duration", "5",
		"--camera-control", `{"type":"simple","config":{"horizontal":5}}`)

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected camera control to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_ValidStaticMask(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--static-mask", maskFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected static mask to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_ValidStaticMaskURL(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--static-mask", "https://example.com/mask.png")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected static mask URL to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_DynamicMaskFileNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", tmpFile.Name(),
		"--dynamic-mask", `[{"mask":"/nonexistent/mask.png","trajectories":[[{"x":100,"y":100}]]}]`)

	if cmdErr == nil {
		t.Fatal("expected error for dynamic mask file not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "mask_not_found" {
		t.Errorf("expected error code 'mask_not_found', got: %s", errorObj["code"])
	}
}

func TestImage2Video_ValidDynamicMaskLocalFile(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--dynamic-mask", `[{"mask":"`+maskFile.Name()+`","trajectories":[[{"x":100,"y":100}]]}]`)

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected dynamic mask local file to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_ValidDynamicMask(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--dynamic-mask", `[{"mask":"https://example.com/mask.png","trajectories":[[{"x":100,"y":100}]]}]`)

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected dynamic mask to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_ValidVoice(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--model", "kling-v2-6",
		"--voice", "voice1,voice2")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected voice to be valid, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_AllFlags(t *testing.T) {
	cmd := newImage2VideoCmd()

	flags := []string{
		"first-frame", "last-frame", "negative", "model", "mode",
		"duration", "cfg-scale", "camera-control", "static-mask",
		"dynamic-mask", "voice", "sound", "watermark", "prompt-file",
	}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestImage2Video_DefaultValues(t *testing.T) {
	cmd := newImage2VideoCmd()

	if cmd.Flag("model").DefValue != "kling-v1" {
		t.Errorf("expected default model 'kling-v1', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("mode").DefValue != "std" {
		t.Errorf("expected default mode 'std', got: %s", cmd.Flag("mode").DefValue)
	}
	if cmd.Flag("duration").DefValue != "5" {
		t.Errorf("expected default duration '5', got: %s", cmd.Flag("duration").DefValue)
	}
}

func TestImage2Video_ShortFlags(t *testing.T) {
	cmd := newImage2VideoCmd()

	shortFlags := map[string]string{
		"i": "first-frame",
		"m": "model",
		"d": "duration",
		"f": "prompt-file",
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

// ===== Image2Video Compatibility Check Tests =====

func TestImage2Video_LastFrameIncompatibleDuration(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	lastFile, err := os.CreateTemp("", "last_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lastFile.Name())
	lastFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--last-frame", lastFile.Name(),
		"--model", "kling-v1",
		"--duration", "10")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible last frame duration")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_last_frame" {
		t.Errorf("expected error code 'incompatible_last_frame', got: %s", errorObj["code"])
	}
}

func TestImage2Video_LastFrameIncompatibleMode(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	lastFile, err := os.CreateTemp("", "last_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lastFile.Name())
	lastFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--last-frame", lastFile.Name(),
		"--model", "kling-v1-5",
		"--mode", "std")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible last frame mode")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_last_frame" {
		t.Errorf("expected error code 'incompatible_last_frame', got: %s", errorObj["code"])
	}
}

func TestImage2Video_MotionBrushIncompatibleModel(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--static-mask", maskFile.Name(),
		"--model", "kling-v2-6")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible motion brush")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_motion_brush" {
		t.Errorf("expected error code 'incompatible_motion_brush', got: %s", errorObj["code"])
	}
}

func TestImage2Video_CameraControlIncompatibleModel(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--camera-control", `{"type":"simple"}`,
		"--model", "kling-v1")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible camera control")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_camera_control" {
		t.Errorf("expected error code 'incompatible_camera_control', got: %s", errorObj["code"])
	}
}

func TestImage2Video_VoiceIncompatibleModel(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--voice", "voice1",
		"--model", "kling-v1")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible voice")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_voice" {
		t.Errorf("expected error code 'incompatible_voice', got: %s", errorObj["code"])
	}
}

func TestImage2Video_LastFrameIncompatibleModel(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	lastFile, err := os.CreateTemp("", "last_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lastFile.Name())
	lastFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--last-frame", lastFile.Name(),
		"--model", "kling-v2-master")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible last frame model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_last_frame" {
		t.Errorf("expected error code 'incompatible_last_frame', got: %s", errorObj["code"])
	}
}

func TestImage2Video_LastFrameCompatible(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	lastFile, err := os.CreateTemp("", "last_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lastFile.Name())
	lastFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--last-frame", lastFile.Name(),
		"--model", "kling-v1",
		"--duration", "5")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected last frame to be compatible, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_MotionBrushIncompatibleDuration(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--static-mask", maskFile.Name(),
		"--model", "kling-v1",
		"--duration", "10")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible motion brush duration")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_motion_brush" {
		t.Errorf("expected error code 'incompatible_motion_brush', got: %s", errorObj["code"])
	}
}

func TestImage2Video_MotionBrushIncompatibleMode(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--static-mask", maskFile.Name(),
		"--model", "kling-v1-5",
		"--mode", "std")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible motion brush mode")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_motion_brush" {
		t.Errorf("expected error code 'incompatible_motion_brush', got: %s", errorObj["code"])
	}
}

func TestImage2Video_MotionBrushIncompatibleV15Duration(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--static-mask", maskFile.Name(),
		"--model", "kling-v1-5",
		"--mode", "pro",
		"--duration", "10")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible motion brush duration")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_motion_brush" {
		t.Errorf("expected error code 'incompatible_motion_brush', got: %s", errorObj["code"])
	}
}

func TestImage2Video_MotionBrushCompatible(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	maskFile, err := os.CreateTemp("", "mask_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(maskFile.Name())
	maskFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--static-mask", maskFile.Name(),
		"--model", "kling-v1",
		"--duration", "5")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected motion brush to be compatible, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_CameraControlIncompatibleMode(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--camera-control", `{"type":"simple"}`,
		"--model", "kling-v1-5",
		"--mode", "std",
		"--duration", "5")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible camera control mode")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_camera_control" {
		t.Errorf("expected error code 'incompatible_camera_control', got: %s", errorObj["code"])
	}
}

func TestImage2Video_CameraControlIncompatibleDuration(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--camera-control", `{"type":"simple"}`,
		"--model", "kling-v1-5",
		"--mode", "pro",
		"--duration", "10")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible camera control duration")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_camera_control" {
		t.Errorf("expected error code 'incompatible_camera_control', got: %s", errorObj["code"])
	}
}

func TestImage2Video_SoundIncompatibleModel(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--sound",
		"--model", "kling-v1")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible sound")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_sound" {
		t.Errorf("expected error code 'incompatible_sound', got: %s", errorObj["code"])
	}
}

func TestImage2Video_SoundCompatible(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--sound",
		"--model", "kling-v2-6")

	if cmdErr == nil {
		t.Fatal("expected error (missing api key)")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected sound to be compatible, got error: %s", errorObj["code"])
	}
}

func TestImage2Video_DynamicMaskIncompatibleModel(t *testing.T) {
	frameFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(frameFile.Name())
	frameFile.Close()

	cmd := NewCmd()
	_, stderr, cmdErr := executeCommand(cmd, "create-from-image",
		"-i", frameFile.Name(),
		"--dynamic-mask", `[{"mask":"https://example.com/mask.png","trajectories":[[{"x":100,"y":100}]]}]`,
		"--model", "kling-v2-6")

	if cmdErr == nil {
		t.Fatal("expected error for incompatible dynamic mask")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "incompatible_motion_brush" {
		t.Errorf("expected error code 'incompatible_motion_brush', got: %s", errorObj["code"])
	}
}
