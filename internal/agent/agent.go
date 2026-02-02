package agent

import (
	"context"
	"strings"
)

type BuildPlan struct {
	Steps []string `json:"steps"`
	Bin   string   `json:"bin"`
}

type Synthesizer interface {
	GenerateBuildPlan(ctx context.Context, repoURL string, files []string, readme string) (*BuildPlan, error)
	SummarizeUpdates(ctx context.Context, commits []string) (string, error)
}

// MockSynthesizer is a temporary implementation for testing.
type MockSynthesizer struct{}

func (m *MockSynthesizer) GenerateBuildPlan(ctx context.Context, repoURL string, files []string, readme string) (*BuildPlan, error) {
	if strings.Contains(repoURL, "anyisland") {
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
