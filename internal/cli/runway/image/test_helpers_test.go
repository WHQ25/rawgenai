package image

import (
	"bytes"
	"os"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "image",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newDownloadCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

func setupNoConfigEnv(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("RUNWAY_API_KEY", "")
}

func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := tmpDir + "/" + name
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return path
}
