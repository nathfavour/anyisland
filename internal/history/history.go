package history

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/pal"
)

type HistoryManager struct {
	sys   pal.System
	agent agent.Synthesizer
}

func NewManager(sys pal.System, ag agent.Synthesizer) *HistoryManager {
	return &HistoryManager{
		sys:   sys,
		agent: ag,
	}
}

func (h *HistoryManager) SyncCommand(ctx context.Context, command string) error {
	// 1. Redact via AI Privacy Firewall
	redacted, err := h.agent.RedactCommand(ctx, command)
	if err != nil {
		return fmt.Errorf("redaction failed: %w", err)
	}

	// 2. Encryption (Mock: prefix with [ENCRYPTED])
	// In a real implementation, we would use AES-GCM or Age here.
	encrypted := fmt.Sprintf("[ENCRYPTED:%s]", redacted)

	// 3. Store in data/history.log
	historyPath := filepath.Join(h.sys.GetDataDir(), "history.log")
	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s | %s\n", time.Now().Format(time.RFC3339), encrypted)
	if _, err := f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func (h *HistoryManager) GetHistory() ([]string, error) {
	historyPath := filepath.Join(h.sys.GetDataDir(), "history.log")
	content, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	
	// Just return the raw lines for now
	return []string{string(content)}, nil
}
