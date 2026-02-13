package agent

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockSynthesizer struct {
	mock.Mock
}

func (m *MockSynthesizer) GenerateBuildPlan(ctx context.Context, repoURL string, files []string, readme string) (*BuildPlan, error) {
	args := m.Called(ctx, repoURL, files, readme)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BuildPlan), args.Error(1)
}

func (m *MockSynthesizer) SummarizeUpdates(ctx context.Context, commits []string) (string, error) {
	args := m.Called(ctx, commits)
	return args.String(0), args.Error(1)
}

func (m *MockSynthesizer) RedactCommand(ctx context.Context, command string) (string, error) {
	args := m.Called(ctx, command)
	return args.String(0), args.Error(1)
}

func (m *MockSynthesizer) AnalyzeDiscretion(ctx context.Context, files []string, readme string) (*DiscretionResult, error) {
	args := m.Called(ctx, files, readme)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DiscretionResult), args.Error(1)
}

func (m *MockSynthesizer) DebugBuildFailure(ctx context.Context, log string, manifest interface{}) (string, error) {
	args := m.Called(ctx, log, manifest)
	return args.String(0), args.Error(1)
}

func (m *MockSynthesizer) DiscoverTool(ctx context.Context, query string) (string, error) {
	args := m.Called(ctx, query)
	return args.String(0), args.Error(1)
}

func (m *MockSynthesizer) ExplainTool(ctx context.Context, name string, manifest interface{}, readme string) (string, error) {
	args := m.Called(ctx, name, manifest, readme)
	return args.String(0), args.Error(1)
}

func (m *MockSynthesizer) SelectAsset(ctx context.Context, assets []string, goos, goarch string) (string, error) {
	args := m.Called(ctx, assets, goos, goarch)
	return args.String(0), args.Error(1)
}
