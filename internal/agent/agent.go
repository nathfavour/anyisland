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
	AnalyzeDiscretion(ctx context.Context, files []string, readme string) (*DiscretionResult, error)
	DebugBuildFailure(ctx context.Context, log string, manifest interface{}) (string, error)
	DiscoverTool(ctx context.Context, query string) (string, error)
	ExplainTool(ctx context.Context, name string, manifest interface{}, readme string) (string, error)
}

type DiscretionResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// HeuristicSynthesizer is a local fallback implementation.

type HeuristicSynthesizer struct{}



func (m *HeuristicSynthesizer) RedactCommand(ctx context.Context, command string) (string, error) {

	// Simple local redaction

	if strings.Contains(command, "KEY") || strings.Contains(command, "PASSWORD") || strings.Contains(command, "SECRET") {

		parts := strings.SplitN(command, "=", 2)

		if len(parts) == 2 {

			return parts[0] + "=[REDACTED]", nil

		}

	}

	return command, nil

}



func (m *HeuristicSynthesizer) GenerateBuildPlan(ctx context.Context, repoURL string, files []string, readme string) (*BuildPlan, error) {

	// If we've reached here, first-class detection in Ingestor failed.

	// We'll provide a very basic generic fallback.

	return &BuildPlan{

		Steps: []string{"# please manually edit anyisland.json - heuristic could not determine build steps"},

		Bin:   "tool",

	}, nil

}



func (m *HeuristicSynthesizer) SummarizeUpdates(ctx context.Context, commits []string) (string, error) {

	if len(commits) == 0 {

		return "No changes.", nil

	}

	return strings.Join(commits, "\n"), nil

}



func (m *HeuristicSynthesizer) AnalyzeDiscretion(ctx context.Context, files []string, readme string) (*DiscretionResult, error) {

	// Use the existing local discretion logic

	res := AnalyzeDiscretion(files, readme)

	return &res, nil

}



func (m *HeuristicSynthesizer) DebugBuildFailure(ctx context.Context, log string, manifest interface{}) (string, error) {



	return "Local Analysis: Build failed. Check the error log above for toolchain issues or missing dependencies.", nil



}







func (m *HeuristicSynthesizer) DiscoverTool(ctx context.Context, query string) (string, error) {



	return "", fmt.Errorf("AI Discovery is unavailable. Please provide a full GitHub URL.")



}







func (m *HeuristicSynthesizer) ExplainTool(ctx context.Context, name string, manifest interface{}, readme string) (string, error) {



	return fmt.Sprintf("Tool: %s. Documentation is available in the source repository.", name), nil



}






