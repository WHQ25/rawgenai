package grok

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

func TestImage_MissingPrompt(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "-o", "output.png")

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

func TestImage_MissingOutput(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat")

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

func TestImage_UnsupportedFormat(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.gif")

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

func TestImage_InvalidN(t *testing.T) {
	tests := []struct {
		name string
		n    string
	}{
		{"zero", "0"},
		{"too large", "11"},
		{"negative", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "-n", tt.n)

			if err == nil {
				t.Fatal("expected error for invalid n")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_n" {
				t.Errorf("expected error code 'invalid_n', got: %s", errorObj["code"])
			}
		})
	}
}

func TestImage_InvalidAspect(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "-a", "2:1")

	if err == nil {
		t.Fatal("expected error for invalid aspect")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_aspect" {
		t.Errorf("expected error code 'invalid_aspect', got: %s", errorObj["code"])
	}
}

func TestImage_ImageNotFound(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Make it better", "-i", "/nonexistent/image.png", "-o", "output.png")

	if err == nil {
		t.Fatal("expected error for image not found")
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

func TestImage_MissingAPIKey(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png")

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

func TestImage_ValidFlags(t *testing.T) {
	cmd := newImageCmd()

	flags := []string{"output", "prompt-file", "image", "n", "aspect"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestImage_DefaultValues(t *testing.T) {
	cmd := newImageCmd()

	if cmd.Flag("n").DefValue != "1" {
		t.Errorf("expected default n '1', got: %s", cmd.Flag("n").DefValue)
	}
	if cmd.Flag("aspect").DefValue != "1:1" {
		t.Errorf("expected default aspect '1:1', got: %s", cmd.Flag("aspect").DefValue)
	}
}

func TestImage_ShortFlags(t *testing.T) {
	cmd := newImageCmd()

	shortFlags := map[string]string{
		"o": "output",
		"i": "image",
		"n": "n",
		"a": "aspect",
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

func TestImage_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("A beautiful landscape")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	t.Setenv("XAI_API_KEY", "")

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", tmpFile.Name(), "-o", "out.png")

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

func TestImage_FromStdin(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")

	cmd := newImageCmd()
	cmd.SetIn(strings.NewReader("A beautiful landscape"))

	_, stderr, err := executeCommand(cmd, "-o", "out.png")

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

func TestImage_EditModeWithImage(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "image_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Make it better", "-i", tmpFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (edit mode validated), got: %s", errorObj["code"])
	}
}

func TestImage_SupportedFormats(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")

	formats := []string{".png", ".jpeg", ".jpg"}
	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out"+ext)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected format '%s' to be supported, got error: %s", ext, errorObj["code"])
			}
		})
	}
}

func TestImage_ValidAspectRatios(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")

	aspects := []string{"1:1", "16:9", "9:16", "4:3", "3:4"}
	for _, aspect := range aspects {
		t.Run(aspect, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "-a", aspect)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected aspect '%s' to be valid, got error: %s", aspect, errorObj["code"])
			}
		})
	}
}

func TestImage_ValidNRange(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")

	for _, n := range []string{"1", "5", "10"} {
		t.Run("n="+n, func(t *testing.T) {
			cmd := newImageCmd()
			_, stderr, err := executeCommand(cmd, "A cute cat", "-o", "out.png", "-n", n)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected n=%s to be valid, got error: %s", n, errorObj["code"])
			}
		})
	}
}

func TestImage_PromptFileNotFound(t *testing.T) {
	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", "/nonexistent/prompt.txt", "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for prompt file not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_EmptyPromptFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "prompt_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	// Write empty content (just whitespace)
	_, _ = tmpFile.WriteString("   \n\t  ")
	tmpFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "--prompt-file", tmpFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for empty prompt file")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestImage_EditModeWithInvalidImageFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "image_*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newImageCmd()
	_, stderr, err := executeCommand(cmd, "Make it better", "-i", tmpFile.Name(), "-o", "out.png")

	if err == nil {
		t.Fatal("expected error for invalid image format")
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
