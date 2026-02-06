package agent

import (
	"context"
	"strings"
)

type BuildPlan struct {
	Steps      []string `json:"steps"`
	Bin        string   `json:"bin"`
	InstallDir string   `json:"install_dir,omitempty"` // Preferred binary path location
	Toolchain  string   `json:"toolchain,omitempty"`   // e.g., "go", "rust", "node"
}

type Synthesizer interface {
	GenerateBuildPlan(ctx context.Context, repoURL string, files []string, readme string) (*BuildPlan, error)
	SummarizeUpdates(ctx context.Context, commits []string) (string, error)
	RedactCommand(ctx context.Context, command string) (string, error)
}

// MockSynthesizer is a temporary implementation for testing.
type MockSynthesizer struct{}

func (m *MockSynthesizer) RedactCommand(ctx context.Context, command string) (string, error) {
	// Simple mock redaction: hide anything after an equals sign if it looks like a secret
	if strings.Contains(command, "KEY") || strings.Contains(command, "PASSWORD") || strings.Contains(command, "SECRET") {
		parts := strings.SplitN(command, "=", 2)
		if len(parts) == 2 {
			return parts[0] + "=[REDACTED]", nil
		}
	}
	return command, nil
}


func (m *MockSynthesizer) GenerateBuildPlan(ctx context.Context, repoURL string, files []string, readme string) (*BuildPlan, error) {
	if strings.Contains(repoURL, "anyisland") || repoURL == "." {
		return &BuildPlan{
			Steps: []string{"go build -o anyisland ./cmd/anyisland"},
			Bin:   "anyisland",
		}, nil
	}
	return &BuildPlan{
		Steps: []string{"go build -o tool ."},
		Bin:   "tool",
	}, nil
}


func (m *MockSynthesizer) SummarizeUpdates(ctx context.Context, commits []string) (string, error) {
	return "Fixed some bugs and added new features.", nil
}
