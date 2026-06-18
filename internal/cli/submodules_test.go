package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSubmoduleTargets(t *testing.T) {
	tmp := t.TempDir()
	gitmodules := `[submodule "anyisland"]
	path = anyisland
	url = https://example.com/anyisland
[submodule "auracrab"]
	path = auracrab
	url = https://example.com/auracrab
`
	if err := os.WriteFile(filepath.Join(tmp, ".gitmodules"), []byte(gitmodules), 0o644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tmp, "anyisland")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"anyisland","version":"1.0.0","build":{"steps":["echo ok"],"bin":"anyisland"}}`
	if err := os.WriteFile(filepath.Join(subDir, "anyisland.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	goDir := filepath.Join(tmp, "auracrab")
	if err := os.MkdirAll(filepath.Join(goDir, "cmd", "auracrab"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module example.com/auracrab\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ing := &Ingestor{}
	targets, err := ing.resolveSubmoduleTargets(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestTrackSubmodules(t *testing.T) {
	ing := &Ingestor{}
	if ing.trackSubmodules(&Manifest{}) {
		t.Fatal("expected false without features")
	}
	if !ing.trackSubmodules(&Manifest{Features: &FeatureConfig{TrackSubmodules: true}}) {
		t.Fatal("expected true with track_submodules")
	}
}
