package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/pal"
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
	oldPath := anyislandPath + ".old"

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

func (m *LifecycleManager) HotSwap() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	m.LogEvent(LifecycleEvent{
		Type:    "update",
		Action:  "hot_swap",
		Status:  "success",
		Message: "Performing hot-swap/exec",
	})

	// In a real scenario, we might pass a --resume-state flag
	return syscallExec(exePath, os.Args, os.Environ())
}

func (m *LifecycleManager) CheckAnyislandUpdate(ctx context.Context, ag agent.Synthesizer) (string, bool, error) {
	ingestor := NewIngestor(ag, m.sys)
	repoURL := "https://github.com/nathfavour/anyisland"
	
	latestCommit, err := ingestor.DiscoverLatestCommit(ctx, repoURL)
	if err != nil {
		return "", false, err
	}

	if latestCommit == Commit && Commit != "none" {
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
	// 1. Fast check
	latest, available, err := m.CheckAnyislandUpdate(ctx, ag)
	if err != nil || !available {
		return
	}

	fmt.Printf("🚀 New version found (%s). Updating Anyisland...\n", latest[:7])

	// 2. Silent update
	ingestor := NewIngestor(ag, m.sys)
	manifest, _, err := ingestor.Ingest(ctx, "https://github.com/nathfavour/anyisland")
	if err != nil {
		fmt.Printf("⚠️ Auto-update ingestion failed: %v\n", err)
		return
	}

	// Build and install the latest version
	_, _, err = ingestor.Build(ctx, manifest, "https://github.com/nathfavour/anyisland")
	if err != nil {
		fmt.Printf("⚠️ Auto-update build failed: %v\n", err)
		return
	}
	
	fmt.Printf("✨ Anyisland auto-updated to %s. Restarting...\n", latest[:7])
	
	// HotSwap to the new binary and re-run the original command
	if err := m.HotSwap(); err != nil {
		fmt.Printf("⚠️ Failed to restart after update: %v\n", err)
		return
	}
	os.Exit(0)
}

func syscallExec(argv0 string, argv []string, envv []string) error {
	return syscall.Exec(argv0, argv, envv)
}