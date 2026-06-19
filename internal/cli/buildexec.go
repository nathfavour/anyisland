package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runBuildStep executes a manifest build step in workDir.
// Shell-style steps (env prefixes, make, quoted flags) run under bash -c.
func runBuildStep(ctx context.Context, workDir, step, toolchain string) error {
	step = strings.TrimSpace(step)
	if step == "" || strings.HasPrefix(step, "#") {
		return nil
	}

	fmt.Printf("Executing: %s\n", step)

	if toolchain == "shell" || needsShell(step) {
		cmd := exec.CommandContext(ctx, "bash", "-c", step)
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("shell step failed: %w", err)
		}
		return nil
	}

	fullArgs := strings.Fields(step)
	if len(fullArgs) == 0 {
		return nil
	}

	var env []string
	args := fullArgs
	for len(args) > 0 && strings.Contains(args[0], "=") && !strings.HasPrefix(args[0], "-") {
		env = append(env, args[0])
		args = args[1:]
	}
	if len(args) == 0 {
		return fmt.Errorf("build step has env vars but no command: %s", step)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = workDir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build step failed: %w", err)
	}
	return nil
}

func needsShell(step string) bool {
	if strings.Contains(step, "&&") || strings.Contains(step, "|") || strings.Contains(step, ";") {
		return true
	}
	fields := strings.Fields(step)
	for i, f := range fields {
		if strings.Contains(f, "=") && !strings.HasPrefix(f, "-") && i == 0 {
			return true
		}
		if f == "make" {
			return true
		}
	}
	for _, f := range fields {
		if strings.Contains(f, "\"") || strings.Contains(f, "'") {
			return true
		}
	}
	return false
}
