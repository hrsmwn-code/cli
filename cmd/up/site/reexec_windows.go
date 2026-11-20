//go:build windows

package site

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fosrl/cli/internal/svcmgr"
)

// reexec restarts the site tunnel. Windows has no execve, so we start a
// detached child process preserving the original arguments (including "up
// site ...") and environment, then exit.
//
// Under the Windows service host a detached child would run alongside the
// replacement the host starts, so in that case we just exit and let the
// host relaunch us.
func reexec() error {
	if supervisedByServiceHost() {
		os.Exit(svcmgr.RestartExitCode)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new process: %w", err)
	}
	os.Exit(0)
	return nil // unreachable
}

func supervisedByServiceHost() bool {
	return os.Getenv(svcmgr.SupervisedEnvVar) != ""
}
