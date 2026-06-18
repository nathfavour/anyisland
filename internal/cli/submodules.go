package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nathfavour/anyisland/internal/agent"
)

func (i *Ingestor) trackSubmodules(m *Manifest) bool {
	return m.Features != nil && m.Features.TrackSubmodules
}

func (i *Ingestor) cloneArgs(repoURL, workDir string, m *Manifest) []string {
	if i.trackSubmodules(m) {
		if i.cfg.Install.DefaultBranch != "" {
			return []string{"clone", "--recursive", "-b", i.cfg.Install.DefaultBranch, repoURL, workDir}
		}
		return []string{"clone", "--recursive", repoURL, workDir}
	}
	if i.cfg.Install.DefaultBranch != "" {
		return []string{"clone", "-b", i.cfg.Install.DefaultBranch, repoURL, workDir}
	}
	return []string{"clone", repoURL, workDir}
}

func (i *Ingestor) syncSubmodules(ctx context.Context, workDir string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", workDir, "submodule", "update", "--init", "--recursive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type submoduleTarget struct {
	Name     string
	WorkDir  string
	Manifest *Manifest
}

func buildPlanForGoModule(name, dir string) agent.BuildPlan {
	target := filepath.Join("cmd", name)
	if _, err := os.Stat(filepath.Join(dir, target)); err != nil {
		target = "."
	} else {
		target = "./" + target
	}
	return agent.BuildPlan{
		Steps:     []string{fmt.Sprintf("go build -ldflags \"-s -w\" -o %s %s", name, target)},
		Bin:       name,
		Toolchain: "go",
	}
}

func (i *Ingestor) resolveSubmoduleTargets(workDir string) ([]submoduleTarget, error) {
	gitmodules := filepath.Join(workDir, ".gitmodules")
	data, err := os.ReadFile(gitmodules)
	if err != nil {
		return nil, nil
	}

	var targets []submoduleTarget
	lines := strings.Split(string(data), "\n")
	var currentPath string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "path = ") {
			currentPath = strings.TrimSpace(strings.TrimPrefix(line, "path = "))
			continue
		}
		if currentPath == "" {
			continue
		}

		subDir := filepath.Join(workDir, currentPath)
		manifestPath := filepath.Join(subDir, "anyisland.json")
		if _, err := os.Stat(manifestPath); err == nil {
			m, err := LoadManifest(manifestPath)
			if err == nil {
				targets = append(targets, submoduleTarget{
					Name:     m.Name,
					WorkDir:  subDir,
					Manifest: m,
				})
			}
		} else if _, err := os.Stat(filepath.Join(subDir, "go.mod")); err == nil {
			name := filepath.Base(currentPath)
			targets = append(targets, submoduleTarget{
				Name:    name,
				WorkDir: subDir,
				Manifest: &Manifest{
					Name:  name,
					Build: buildPlanForGoModule(name, subDir),
				},
			})
		}
		currentPath = ""
	}
	return targets, nil
}

func (i *Ingestor) buildSubmoduleTargets(ctx context.Context, workDir string, m *Manifest) error {
	if !i.trackSubmodules(m) {
		return nil
	}

	if err := i.syncSubmodules(ctx, workDir); err != nil {
		return fmt.Errorf("submodule sync failed: %w", err)
	}

	targets, err := i.resolveSubmoduleTargets(workDir)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	binDir := i.sys.GetIslandBinDir()
	if m.InstallDir != "" {
		binDir = i.ExpandPath(m.InstallDir)
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(targets))

	for _, target := range targets {
		wg.Add(1)
		go func(t submoduleTarget) {
			defer wg.Done()
			if err := i.buildSubmoduleBinary(ctx, t, binDir); err != nil {
				errCh <- fmt.Errorf("%s: %w", t.Name, err)
			}
		}(target)
	}

	wg.Wait()
	close(errCh)

	var errs []string
	for err := range errCh {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf("submodule builds failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (i *Ingestor) buildSubmoduleBinary(ctx context.Context, target submoduleTarget, binDir string) error {
	plan := target.Manifest.Build
	for _, step := range plan.Steps {
		step = strings.TrimSpace(step)
		if step == "" {
			continue
		}
		args := strings.Fields(step)
		if len(args) == 0 {
			continue
		}
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = target.WorkDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build step failed: %w", err)
		}
	}

	src := filepath.Join(target.WorkDir, plan.Bin)
	dst := filepath.Join(binDir, filepath.Base(plan.Bin))
	if target.Manifest.BinName != "" {
		dst = filepath.Join(binDir, target.Manifest.BinName)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		return err
	}
	fmt.Printf("Installed submodule binary %s -> %s\n", target.Name, dst)
	return nil
}
