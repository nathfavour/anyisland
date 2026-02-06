package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
