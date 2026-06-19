package packagemanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/nathfavour/anyisland/pkg/ipc"
)

// DefaultDistributor publishes binaries into the local anyisland tree.
type DefaultDistributor struct {
	BinDir string
}

func NewDefaultDistributor(binDir string) *DefaultDistributor {
	if binDir == "" {
		binDir = ipc.BinDir()
	}
	return &DefaultDistributor{BinDir: binDir}
}

func (d *DefaultDistributor) Distribute(ctx context.Context, req DistributionRequest) (*DistributionResult, error) {
	if req.BinPath == "" {
		return nil, fmt.Errorf("bin path is required")
	}
	if req.Package == "" {
		return nil, fmt.Errorf("package name is required")
	}

	src, err := filepath.Abs(req.BinPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(src); err != nil {
		return nil, fmt.Errorf("binary not found: %w", err)
	}

	if err := os.MkdirAll(d.BinDir, 0o755); err != nil {
		return nil, err
	}

	name := filepath.Base(src)
	if req.Package != "" {
		name = filepath.Base(req.Package)
	}
	dest := filepath.Join(d.BinDir, name)

	if err := copyFile(src, dest); err != nil {
		return nil, err
	}
	if err := os.Chmod(dest, 0o755); err != nil {
		return nil, err
	}

	hash, err := hashFile(dest)
	if err != nil {
		return nil, err
	}

	_ = notifyAnyisland(req.Package)

	manifest := map[string]interface{}{
		"package": req.Package,
		"version": req.Version,
		"sha256":  hash,
		"path":    dest,
		"room_id": req.RoomID,
	}

	return &DistributionResult{
		Success:     true,
		Hash:        hash,
		DownloadURL: "file://" + dest,
		Manifest:    manifest,
	}, nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func notifyAnyisland(tool string) error {
	conn, err := net.DialTimeout("unix", ipc.SocketPath(), 2*time.Second)
	if err != nil {
		return nil
	}
	defer conn.Close()
	req := map[string]string{"op": "QUERY", "tool": tool}
	_ = json.NewEncoder(conn).Encode(req)
	return nil
}
