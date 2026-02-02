package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/pal"
)

type Ingestor struct {
	gh    *github.Client
	agent agent.Synthesizer
	sys   pal.System
}

func NewIngestor(ag agent.Synthesizer, sys pal.System) *Ingestor {
	return &Ingestor{
		gh:    github.NewClient(nil),
		agent: ag,
		sys:   sys,
	}
}

func (i *Ingestor) Build(ctx context.Context, plan *agent.BuildPlan, repoURL string) error {
	parts := strings.Split(strings.TrimPrefix(repoURL, "https://"), "/")
	repoName := parts[len(parts)-1]
	cloneDir := filepath.Join(i.sys.GetCacheDir(), repoName)

	// 1. Clone or Copy
	fmt.Printf("Fetching %s...\n", repoURL)
	if err := os.RemoveAll(cloneDir); err != nil {
		return err
	}

	source := repoURL
	if !strings.HasPrefix(source, "http") && !strings.HasPrefix(source, "git@") {
		// Assume local if it doesn't look like a URL
		if _, err := os.Stat(source); err == nil {
			source, _ = filepath.Abs(source)
		} else {
			source = "https://" + source
		}
	}

	cmd := exec.CommandContext(ctx, "git", "clone", source, cloneDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed from %s: %w", source, err)
	}


	// 2. Execute build steps
	for _, step := range plan.Steps {
		fmt.Printf("Executing: %s\n", step)
		args := strings.Fields(step)
		buildCmd := exec.CommandContext(ctx, args[0], args[1:]...)
		buildCmd.Dir = cloneDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("build step failed: %w", err)
		}
	}

	// 3. Move binary
	srcBin := filepath.Join(cloneDir, plan.Bin)
	dstBin := filepath.Join(i.sys.GetBinDir(), plan.Bin)
	fmt.Printf("Installing %s to %s...\n", srcBin, dstBin)
	
	// Copy file (more robust than Move across filesystems)
	input, err := os.ReadFile(srcBin)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dstBin, input, 0755); err != nil {
		return err
	}

	return nil
}


func (i *Ingestor) Ingest(ctx context.Context, repoURL string) (*agent.BuildPlan, error) {
	// Simple parsing of "github.com/user/repo"
	parts := strings.Split(strings.TrimPrefix(repoURL, "https://"), "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid github URL: %s", repoURL)
	}
	owner := parts[1]
	repo := parts[2]

	fmt.Printf("Fetching repository info for %s/%s...\n", owner, repo)

	// Fetch file tree (simplified)
	tree, _, err := i.gh.Git.GetTree(ctx, owner, repo, "main", true)
	if err != nil {
		// Try master if main fails
		tree, _, err = i.gh.Git.GetTree(ctx, owner, repo, "master", true)
		if err != nil {
			return nil, err
		}
	}

	var files []string
	for _, entry := range tree.Entries {
		files = append(files, entry.GetPath())
	}

	// Fetch README
	readme, _, err := i.gh.Repositories.GetReadme(ctx, owner, repo, nil)
	readmeContent := ""
	if err == nil {
		content, _ := readme.GetContent()
		readmeContent = content
	}

	fmt.Println("Generating build plan via AI...")
	return i.agent.GenerateBuildPlan(ctx, repoURL, files, readmeContent)
}
