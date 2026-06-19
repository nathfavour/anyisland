package packagemanager

import "context"

// DistributionRequest carries release metadata for publishing.
type DistributionRequest struct {
	Package  string
	Version  string
	WorkDir  string
	BinPath  string
	RoomID   string
}

// DistributionResult reports signed release metadata.
type DistributionResult struct {
	Success     bool
	Hash        string
	DownloadURL string
	Manifest    map[string]interface{}
}

// DistributionInterface signs, publishes, and records immutable release metadata.
type DistributionInterface interface {
	Distribute(ctx context.Context, req DistributionRequest) (*DistributionResult, error)
}
