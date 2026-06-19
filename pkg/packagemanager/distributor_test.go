package packagemanager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDistributor(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "tool")
	if err := os.WriteFile(src, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	binDir := filepath.Join(tmp, "bin")
	d := NewDefaultDistributor(binDir)
	res, err := d.Distribute(t.Context(), DistributionRequest{
		Package: "tool",
		Version: "1.0.0",
		BinPath: src,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.Hash == "" {
		t.Fatalf("expected successful distribution, got %+v", res)
	}
	if _, err := os.Stat(filepath.Join(binDir, "tool")); err != nil {
		t.Fatal(err)
	}
}
