package cli

import (
	"context"
	"encoding/json"
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
	repoURL = normalizeRepoURL(repoURL)
	var workDir string

	if _, err := os.Stat(repoURL); err == nil {
		// Local path: use it directly
		workDir = repoURL
		fmt.Printf("Using local source at %s\n", workDir)
	} else {
		// Remote: manage in SourceDir
		parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
		if len(parts) < 2 {
			return fmt.Errorf("invalid github URL: %s", repoURL)
		}
		repoPath := filepath.Join(parts...)
		workDir = filepath.Join(i.sys.GetSourceDir(), repoPath)

		if _, err := os.Stat(filepath.Join(workDir, ".git")); os.IsNotExist(err) {
			fmt.Printf("Cloning %s to %s...\n", repoURL, workDir)
			if err := os.MkdirAll(filepath.Dir(workDir), 0755); err != nil {
				return err
			}
			cmd := exec.CommandContext(ctx, "git", "clone", repoURL, workDir)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}
		} else {
			fmt.Printf("Updating %s in %s...\n", repoURL, workDir)
			// Efficient update using fetch + reset --hard to origin/master or main
			fetchCmd := exec.CommandContext(ctx, "git", "-C", workDir, "fetch", "origin")
			if err := fetchCmd.Run(); err != nil {
				return fmt.Errorf("git fetch failed: %w", err)
			}

			// Try to reset to origin/master or origin/main
			resetCmd := exec.CommandContext(ctx, "git", "-C", workDir, "reset", "--hard", "origin/master")
			if err := resetCmd.Run(); err != nil {
				resetCmd = exec.CommandContext(ctx, "git", "-C", workDir, "reset", "--hard", "origin/main")
				if err := resetCmd.Run(); err != nil {
					return fmt.Errorf("git reset failed: %w", err)
				}
			}
		}
	}

	// 2. Execute build steps
	for _, step := range plan.Steps {
		fmt.Printf("Executing: %s\n", step)
		args := strings.Fields(step)
		buildCmd := exec.CommandContext(ctx, args[0], args[1:]...)
		buildCmd.Dir = workDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("build step failed: %w", err)
		}
	}

	// 3. Move binary
	srcBin := filepath.Join(workDir, plan.Bin)
	
	targetDir := i.sys.GetIslandBinDir()
	if plan.Bin == "anyisland" || plan.Bin == "anyislandd" {
		targetDir = i.sys.GetBinDir()
	}
	
	dstBin := filepath.Join(targetDir, plan.Bin)
	fmt.Printf("Installing %s to %s...\n", srcBin, dstBin)
	
	// Handle "text file busy" by renaming existing binary first
	if _, err := os.Stat(dstBin); err == nil {
		oldBin := dstBin + ".old"
		_ = os.Remove(oldBin) // Clean up any previous old binary
		if err := os.Rename(dstBin, oldBin); err != nil {
			return fmt.Errorf("failed to move existing binary: %w", err)
		}
		defer os.Remove(oldBin) // Try to clean up after successful installation
	}

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
	repoURL = normalizeRepoURL(repoURL)
	var owner, repo string
	var files []string
	var readmeContent string

	if _, err := os.Stat(repoURL); err == nil {
		// Local path
		fmt.Printf("Analyzing local repository %s...\n", repoURL)
		absPath, _ := filepath.Abs(repoURL)
		
		manifestPath := filepath.Join(absPath, "anyisland.json")
		if _, err := os.Stat(manifestPath); err == nil {
			fmt.Println("Found anyisland.json, using provided build plan.")
			m, err := LoadManifest(manifestPath)
			if err == nil {
				return &m.Build, nil
			}
			fmt.Printf("Warning: failed to parse anyisland.json: %v\n", err)
		}

		repo = filepath.Base(absPath)
		err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(absPath, path)
			files = append(files, rel)
			if strings.ToLower(info.Name()) == "readme.md" {
				content, _ := os.ReadFile(path)
				readmeContent = string(content)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Remote GitHub URL
		parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid github URL: %s", repoURL)
		}
		owner = parts[0]
		repo = parts[1]

		fmt.Printf("Fetching repository info for %s/%s...\n", owner, repo)

		// Try to fetch anyisland.json first
		fileContent, _, _, err := i.gh.Repositories.GetContents(ctx, owner, repo, "anyisland.json", nil)
		if err == nil {
			content, _ := fileContent.GetContent()
			var m Manifest
			if err := json.Unmarshal([]byte(content), &m); err == nil {
				fmt.Println("Found anyisland.json, using provided build plan.")
				return &m.Build, nil
			}
		}

		// Fetch file tree (simplified)
		tree, _, err := i.gh.Git.GetTree(ctx, owner, repo, "main", true)
		if err != nil {
			// Try master if main fails
			tree, _, err = i.gh.Git.GetTree(ctx, owner, repo, "master", true)
			if err != nil {
				return nil, err
			}
		}

		for _, entry := range tree.Entries {
			files = append(files, entry.GetPath())
		}

		// Fetch README
		readme, _, err := i.gh.Repositories.GetReadme(ctx, owner, repo, nil)
		if err == nil {
			content, _ := readme.GetContent()
			readmeContent = content
		}
	}

	fmt.Println("Generating build plan via AI...")
	return i.agent.GenerateBuildPlan(ctx, repoURL, files, readmeContent)
}

func normalizeRepoURL(url string) string {
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "git@") {
		if _, err := os.Stat(url); err == nil {
			abs, _ := filepath.Abs(url)
			return abs
		}
		if strings.Count(url, "/") == 1 {
			return "https://github.com/" + url
		}
		return "https://" + url
	}
	return url
}

