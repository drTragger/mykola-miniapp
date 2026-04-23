package sysutil

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func RunCommand(timeoutSec int, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timeout")
	}

	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

func RunSudoCommand(timeoutSec int, name string, args ...string) (string, error) {
	allArgs := append([]string{"-n", name}, args...)
	return RunCommand(timeoutSec, "sudo", allArgs...)
}

func IsServiceActive(name string) bool {
	out, err := exec.Command("systemctl", "is-active", name).Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) == "active"
}
