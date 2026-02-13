package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/nathfavour/anyisland/internal/registry"
)

type LifecycleEvent struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Action    string `json:"action"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Version   string `json:"version,omitempty"`
	Commit    string `json:"commit,omitempty"`
}

type LifecycleManager struct {
	sys pal.System
}

func NewLifecycleManager(sys pal.System) *LifecycleManager {
	return &LifecycleManager{sys: sys}
}

func (m *LifecycleManager) LogEvent(evt LifecycleEvent) error {
	evt.Timestamp = time.Now().Format(time.RFC3339)
	auditDir := filepath.Join(m.sys.GetIslandDir(), "audit")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(auditDir, "lifecycle.jsonl")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, string(data))
	return err
}

func (m *LifecycleManager) SelfInstall() error {
	// 1. Ensure folders exist
	if err := m.sys.InitFolders(); err != nil {
		return err
	}

	// 2. Migration
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	binName := filepath.Base(exePath)
	targetPath := filepath.Join(m.sys.GetBinDir(), binName)

	// Don't copy if we are already running from the target path
	if exePath != targetPath {
		data, err := os.ReadFile(exePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, data, 0755); err != nil {
			return err
		}
	}

	// 3. Shell Injection
	if err := m.sys.InjectPath(); err != nil {
		return err
	}

	// 4. Register in database
	reg, err := registry.Open(m.sys.GetIslandDir())
	if err == nil {
		defer reg.Close()
		reg.RegisterTool(registry.Tool{
			Name:        "anyisland",
			Source:      "https://github.com/nathfavour/anyisland",
			SourceDir:   os.Getenv("PWD"), // Assume current dir for self-install
			Version:     Version,
			LastCommit:  Commit,
			InstallPath: targetPath,
			Type:        "source",
		})
	}

	return m.LogEvent(LifecycleEvent{
		Type:    "install",
		Action:  "self_install",
		Status:  "success",
		Message: fmt.Sprintf("Installed %s to %s", binName, targetPath),
		Version: Version,
		Commit:  Commit,
	})
}

func (m *LifecycleManager) Rollback() error {
	binDir := m.sys.GetBinDir()
	anyislandPath := filepath.Join(binDir, "anyisland")
	oldPath := anyislandPath + ".bak"

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("no rollback version found at %s", oldPath)
	}

	// Swap
	tempPath := anyislandPath + ".tmp"
	if err := os.Rename(anyislandPath, tempPath); err != nil {
		return err
	}
	if err := os.Rename(oldPath, anyislandPath); err != nil {
		os.Rename(tempPath, anyislandPath) // attempt recovery
		return err
	}
	os.Remove(tempPath)

	return m.LogEvent(LifecycleEvent{
		Type:    "rollback",
		Action:  "binary_swap",
		Status:  "success",
		Message: "Rolled back to previous version",
	})
}

func (m *LifecycleManager) HealTool(ctx context.Context, ag agent.Synthesizer, toolName string) error {
	reg, err := registry.Open(m.sys.GetIslandDir())
	if err != nil {
		return err
	}
	defer reg.Close()

	t, err := reg.GetTool(toolName)
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("tool %s not found in registry", toolName)
	}

	ingestor := NewIngestor(ag, m.sys, nil)
	isBroken := false
	reason := ""

	// 1. Check binary existence
	if _, err := os.Stat(t.InstallPath); os.IsNotExist(err) {
		isBroken = true
		reason = "binary missing"
	} else if !ingestor.VerifyToolIntegrity(t.InstallPath, t.BinaryHash) {
		// 2. Check integrity (hash)
		isBroken = true
		reason = "integrity mismatch (tampered or corrupted)"
	}

	// 3. Check source directory (if applicable)
	sourceDir := t.SourceDir
	if sourceDir == "" {
		sourceDir = ingestor.getSourcePath(t.Source, t.Name)
	} else {
		sourceDir = ingestor.expandPath(sourceDir)
	}

	if t.Type == "source" {
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			isBroken = true
			reason = "source directory missing"
		} else {
			// Only require .git if it's NOT a local absolute path
			// (meaning it's a cached remote repo)
			isLocal := false
			if filepath.IsAbs(t.Source) {
				if _, err := os.Stat(t.Source); err == nil {
					isLocal = true
				}
			}

			if !isLocal {
				if _, err := os.Stat(filepath.Join(sourceDir, ".git")); os.IsNotExist(err) {
					// Some tools might delete .git to "clean up" but we need it for updates
					isBroken = true
					reason = "source directory corrupted (missing .git)"
				}
			}
		}
	}

	if !isBroken {
		return nil // Tool is healthy
	}

	fmt.Printf("🛠️ Tool '%s' is broken: %s. Healing...\n", toolName, reason)

	m.LogEvent(LifecycleEvent{
		Type:    "heal",
		Action:  "start",
		Status:  "processing",
		Message: fmt.Sprintf("Healing %s due to: %s", toolName, reason),
	})

	// Re-ingest and Re-build
	manifest, commit, _, err := ingestor.Ingest(ctx, t.Source)
	if err != nil {
		return fmt.Errorf("failed to re-ingest during heal: %w", err)
	}

	hash, installPath, err := ingestor.Build(ctx, manifest, t.Source)
	if err != nil {
		return fmt.Errorf("failed to re-build during heal: %w", err)
	}

	// Update registry with new health info
	t.BinaryHash = hash
	t.InstallPath = installPath
	t.Version = manifest.Version
	t.LastCommit = commit
	t.SourceDir = manifest.SourceDir
	if err := reg.RegisterTool(*t); err != nil {
		return fmt.Errorf("failed to update registry after heal: %w", err)
	}

	return m.LogEvent(LifecycleEvent{
		Type:    "heal",
		Action:  "complete",
		Status:  "success",
		Message: fmt.Sprintf("Healed %s successfully", toolName),
	})
}

func (m *LifecycleManager) HotSwap(targetPath string, extraEnv ...string) error {
	exePath := targetPath
	if exePath == "" {
		var err error
		exePath, err = os.Executable()
		if err != nil {
			return err
		}
	}

	m.LogEvent(LifecycleEvent{
		Type:    "update",
		Action:  "hot_swap",
		Status:  "success",
		Message: fmt.Sprintf("Performing hot-swap/exec to %s", exePath),
	})

	env := os.Environ()
	env = append(env, extraEnv...)

	return syscallExec(exePath, os.Args, env)
}

func (m *LifecycleManager) CheckAnyislandUpdate(ctx context.Context, ag agent.Synthesizer) (string, bool, error) {
	ingestor := NewIngestor(ag, m.sys, nil)
	repoURL := "https://github.com/nathfavour/anyisland"
	
	latestCommit, err := ingestor.DiscoverLatestCommit(ctx, repoURL)
	if err != nil {
		return "", false, err
	}

	curCommit := GetEffectiveCommit()
	// Normalize for comparison: strip "(dirty)" suffix if present
	cleanCurCommit := strings.Split(curCommit, " ")[0]

	if latestCommit == cleanCurCommit && cleanCurCommit != "none" {
		return latestCommit, false, nil
	}

	m.LogEvent(LifecycleEvent{
		Type:    "update",
		Action:  "check",
		Status:  "success",
		Message: "New version discovered",
		Commit:  latestCommit,
	})

	return latestCommit, true, nil
}

func (m *LifecycleManager) BackgroundAutoUpdate(ctx context.Context, ag agent.Synthesizer) {
	cm := NewConfigManager(m.sys)
	cfg, _ := cm.Load()
	if !cfg.Update.AutoUpdate {
		return
	}

	// Fast check (uses git ls-remote)
	latest, available, err := m.CheckAnyislandUpdate(ctx, ag)
	if err != nil || !available {
		return
	}

	fmt.Printf("🚀 New version found (%s). Auto-updating Anyisland...\n", ShortCommit(latest))

	// 3. Silent update
	ingestor := NewIngestor(ag, m.sys, nil)
	manifest, _, _, err := ingestor.Ingest(ctx, "https://github.com/nathfavour/anyisland")
	if err != nil {
		fmt.Printf("⚠️ Auto-update ingestion failed: %v\n", err)
		return
	}

	// Build and install the latest version
	hash, installPath, err := ingestor.Build(ctx, manifest, "https://github.com/nathfavour/anyisland")
	if err != nil {
		fmt.Printf("⚠️ Auto-update build failed: %v\n", err)
		return
	}

	// Update registry
	reg, err := registry.Open(m.sys.GetIslandDir())
	if err == nil {
		defer reg.Close()
		reg.RegisterTool(registry.Tool{
			Name:        manifest.Name,
			Source:      "https://github.com/nathfavour/anyisland",
			SourceDir:   manifest.SourceDir,
			Version:     manifest.Version,
			LastCommit:  latest,
			BinaryHash:  hash,
			InstallPath: installPath,
			Type:        "source",
		})
	}
	
	fmt.Printf("✨ Anyisland auto-updated to %s. Restarting...\n", ShortCommit(latest))
	
	// HotSwap to the new binary and re-run the original command
	// We pass a signal that we just updated to avoid redundant "already up to date" messages
	if err := m.HotSwap(installPath, "ANYISLAND_JUST_UPDATED=1"); err != nil {
		fmt.Printf("⚠️ Failed to restart after update: %v\n", err)
		return
	}
	os.Exit(0)
}

func syscallExec(argv0 string, argv []string, envv []string) error {
	return syscall.Exec(argv0, argv, envv)
}