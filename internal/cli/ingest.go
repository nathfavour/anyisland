package cli

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func (i *Ingestor) DiscoverLatestCommit(ctx context.Context, repoURL string) (string, error) {
	repoURL = normalizeRepoURL(repoURL)
	cmd := exec.CommandContext(ctx, "git", "ls-remote", repoURL, "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch remote commit: %w", err)
	}

	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid ls-remote output")
	}

	return parts[0], nil
}

func (i *Ingestor) getSourcePath(repoURL string, pkgName string) string {
	if _, err := os.Stat(repoURL); err == nil {
		abs, _ := filepath.Abs(repoURL)
		return abs
	}

	cleanURL := strings.TrimPrefix(repoURL, "https://")
	cleanURL = strings.TrimPrefix(cleanURL, "http://")
	cleanURL = strings.TrimPrefix(cleanURL, "git@")
	cleanURL = strings.Replace(cleanURL, ":", "/", 1)
	cleanURL = strings.TrimSuffix(cleanURL, ".git")

	parts := strings.Split(cleanURL, "/")
	// Detect provider (github.com, gitlab.com, etc)
	if len(parts) >= 3 {
		host := parts[0]
		user := parts[1]
		repo := parts[2]

		provider := ""
		if strings.Contains(host, "github") {
			provider = "github"
		} else if strings.Contains(host, "gitlab") {
			provider = "gitlab"
		}

		if provider != "" {
			return filepath.Join(i.sys.GetSourceDir(), provider, user, repo)
		}
	}

	return filepath.Join(i.sys.GetSourceDir(), pkgName)
}

func (i *Ingestor) Build(ctx context.Context, m *Manifest, repoURL string) error {
	repoURL = normalizeRepoURL(repoURL)
	workDir := i.getSourcePath(repoURL, m.Name)

	if _, err := os.Stat(repoURL); err == nil {
		// Local path: use it directly
		fmt.Printf("Using local source at %s\n", workDir)
	} else if strings.HasSuffix(repoURL, ".zip") || strings.Contains(repoURL, "/zipball/") || strings.Contains(repoURL, "/archive/") {
		// ZIP source
		if err := i.downloadAndUnzip(ctx, repoURL, workDir); err != nil {
			return err
		}
	} else {
		// Remote Git repo
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
			fetchCmd := exec.CommandContext(ctx, "git", "-C", workDir, "fetch", "origin")
			if err := fetchCmd.Run(); err != nil {
				return fmt.Errorf("git fetch failed: %w", err)
			}

			resetCmd := exec.CommandContext(ctx, "git", "-C", workDir, "reset", "--hard", "origin/master")
			if err := resetCmd.Run(); err != nil {
				resetCmd = exec.CommandContext(ctx, "git", "-C", workDir, "reset", "--hard", "origin/main")
				if err := resetCmd.Run(); err != nil {
					return fmt.Errorf("git reset failed: %w", err)
				}
			}
		}
	}

	// Check if this was actually a git repo (important for zips)
	if _, err := os.Stat(filepath.Join(workDir, ".git")); err == nil {
		gitURLCmd := exec.Command("git", "-C", workDir, "remote", "get-url", "origin")
		remoteURLBytes, err := gitURLCmd.Output()
		if err == nil {
			remoteURL := strings.TrimSpace(string(remoteURLBytes))
			structuredPath := i.getSourcePath(remoteURL, m.Name)
			if structuredPath != workDir {
				fmt.Printf("Detected git repo in source, moving to %s\n", structuredPath)
				os.MkdirAll(filepath.Dir(structuredPath), 0755)
				if err := os.Rename(workDir, structuredPath); err == nil {
					workDir = structuredPath
				}
			}
		}
	}

	// 1.5 Craft anyisland.json if missing
	manifestPath := filepath.Join(workDir, "anyisland.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Println("Crafting anyisland.json for this tool...")
		data, err := json.MarshalIndent(m, "", "  ")
		if err == nil {
			_ = os.WriteFile(manifestPath, data, 0644)
		}
	}

	// 2. Execute build steps

	plan := m.Build

	// Go-specific Pre-build check

	if plan.Toolchain == "go" {

		if _, err := exec.LookPath("go"); err != nil {

			return fmt.Errorf("go toolchain required but not found in PATH")

		}

	}

	// Rust-specific Pre-build check

	if plan.Toolchain == "rust" {

		if _, err := exec.LookPath("cargo"); err != nil {

			return fmt.Errorf("rust toolchain (cargo) required but not found in PATH")

		}

	}

	// Python-specific Pre-build check

	if plan.Toolchain == "python" {

		if _, err := exec.LookPath("python3"); err != nil {

			if _, err := exec.LookPath("python"); err != nil {

				return fmt.Errorf("python3 or python required but not found in PATH")

			}

		}

	}

		for _, step := range plan.Steps {

			fmt.Printf("Executing: %s\n", step)

			args := strings.Fields(step)

			buildCmd := exec.CommandContext(ctx, args[0], args[1:]...)

			buildCmd.Dir = workDir

	

			// Capture output for potential debugging

			var combinedOutput strings.Builder

			buildCmd.Stdout = io.MultiWriter(os.Stdout, &combinedOutput)

			buildCmd.Stderr = io.MultiWriter(os.Stderr, &combinedOutput)

	

			if err := buildCmd.Run(); err != nil {

				fmt.Println("\n❌ Build step failed.")

				if v, ok := i.agent.(*agent.VibeauraSynthesizer); ok {

					fmt.Println("Analyzing failure with AI...")

					analysis, diagErr := v.DebugBuildFailure(ctx, combinedOutput.String(), m)

					if diagErr == nil {

						fmt.Printf("\nAI Analysis:\n%s\n", analysis)

					}

				}

				return fmt.Errorf("build step failed: %w", err)

			}

		}

	

	// 3. Move binary/bundle

	binPattern := filepath.Join(workDir, plan.Bin)

	matches, err := filepath.Glob(binPattern)

	srcBin := ""

	if err == nil && len(matches) > 0 {

		srcBin = matches[0] // Pick the first match

	} else {

		srcBin = binPattern // Fallback to literal path

	}

	targetDir := ""

	if m.InstallDir != "" {
		targetDir = m.InstallDir
	} else if plan.InstallDir != "" {
		targetDir = plan.InstallDir
	} else {
		targetDir = i.sys.GetIslandBinDir()
		if plan.Bin == "anyisland" || plan.Bin == "anyislandd" {
			targetDir = i.sys.GetBinDir()
		}
	}

	if plan.Toolchain == "flutter" {

		// Flutter needs its entire bundle directory to run (data, lib, etc.)

		// We install it to a subdirectory and create a wrapper script

		appDir := filepath.Join(targetDir, m.Name+"-app")

		_ = os.RemoveAll(appDir)

		if err := os.MkdirAll(appDir, 0755); err != nil {

			return fmt.Errorf("failed to create app directory: %w", err)

		}

		fmt.Printf("Deploying Flutter bundle to %s...\n", appDir)

		// Use shell to copy directory contents

		cpCmd := exec.Command("cp", "-r", srcBin+"/.", appDir)

		if runtime.GOOS == "windows" {

			cpCmd = exec.Command("xcopy", "/E", "/I", "/Y", srcBin, appDir)

		}

		if err := cpCmd.Run(); err != nil {

			return fmt.Errorf("failed to copy flutter bundle: %w", err)

		}

		// Find the actual executable in the bundle

		entries, _ := os.ReadDir(appDir)

		exeName := m.Name

		for _, entry := range entries {

			if !entry.IsDir() && (entry.Type().IsRegular() || entry.Type() == 0) {

				// Basic heuristic: the binary usually has the project name or is the only non-extension file

				if !strings.Contains(entry.Name(), ".") {

					exeName = entry.Name()

					break

				}

			}

		}

		// Create wrapper script in the main bin directory

		wrapperPath := filepath.Join(targetDir, m.Name)

		wrapperContent := fmt.Sprintf("#!/bin/bash\ncd %s && ./%s \"$@\"\n", appDir, exeName)

		if runtime.GOOS == "windows" {

			wrapperPath += ".bat"

			wrapperContent = fmt.Sprintf("@echo off\ncd /d %%~dp0\\%s-app\nstart \"\" %s.exe %%*\n", m.Name, exeName)

		}

		if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {

			return fmt.Errorf("failed to create wrapper script: %w", err)

		}

		fmt.Printf("Created wrapper at %s\n", wrapperPath)

		return nil

	}

	if plan.Toolchain == "node" {

		// Node.js projects need their node_modules and package.json

		appDir := filepath.Join(targetDir, m.Name+"-app")

		_ = os.RemoveAll(appDir)

		if err := os.MkdirAll(appDir, 0755); err != nil {

			return fmt.Errorf("failed to create app directory: %w", err)

		}

		fmt.Printf("Deploying Node.js project to %s...\n", appDir)

		cpCmd := exec.Command("cp", "-r", workDir+"/.", appDir)

		if runtime.GOOS == "windows" {

			cpCmd = exec.Command("xcopy", "/E", "/I", "/Y", workDir, appDir)

		}

		if err := cpCmd.Run(); err != nil {

			return fmt.Errorf("failed to copy node project: %w", err)

		}

		// Try to find the entry point in package.json

		pkgData, err := os.ReadFile(filepath.Join(appDir, "package.json"))

		var binScript string

		if err == nil {

			var pkg struct {
				Bin map[string]string `json:"bin"`
			}

			if err := json.Unmarshal(pkgData, &pkg); err == nil && len(pkg.Bin) > 0 {

				// Just pick the first one if multiple are defined

				for _, script := range pkg.Bin {

					binScript = script

					break

				}

			}

		}

		if binScript == "" {

			binScript = "index.js" // fallback

		}

		// Create wrapper script

		wrapperPath := filepath.Join(targetDir, m.Name)

		wrapperContent := fmt.Sprintf("#!/bin/bash\nnode %s/%s \"$@\"\n", appDir, binScript)

		if runtime.GOOS == "windows" {

			wrapperPath += ".bat"

			wrapperContent = fmt.Sprintf("@echo off\nnode %%~dp0\\%s-app\\%s %%*\n", m.Name, binScript)

		}

		if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {

			return fmt.Errorf("failed to create wrapper script: %w", err)

		}

		fmt.Printf("Created wrapper at %s\n", wrapperPath)

		return nil

	}

	if plan.Toolchain == "python" {

		// Python projects live in their venv

		appDir := filepath.Join(targetDir, m.Name+"-app")

		_ = os.RemoveAll(appDir)

		if err := os.MkdirAll(appDir, 0755); err != nil {

			return fmt.Errorf("failed to create app directory: %w", err)

		}

		fmt.Printf("Deploying Python project to %s...\n", appDir)

		cpCmd := exec.Command("cp", "-r", workDir+"/.", appDir)

		if runtime.GOOS == "windows" {

			cpCmd = exec.Command("xcopy", "/E", "/I", "/Y", workDir, appDir)

		}

		if err := cpCmd.Run(); err != nil {

			return fmt.Errorf("failed to copy python project: %w", err)

		}

		// Heuristic to find the binary in venv/bin

		venvBinDir := filepath.Join(appDir, "venv", "bin")

		if runtime.GOOS == "windows" {

			venvBinDir = filepath.Join(appDir, "venv", "Scripts")

		}

		entries, _ := os.ReadDir(venvBinDir)

		binName := m.Name

		// Look for an exact match or something that looks like the main script

		for _, entry := range entries {

			if strings.EqualFold(entry.Name(), m.Name) || strings.EqualFold(entry.Name(), m.Name+".exe") {

				binName = entry.Name()

				break

			}

		}

		// Create wrapper script

		wrapperPath := filepath.Join(targetDir, m.Name)

		wrapperContent := fmt.Sprintf("#!/bin/bash\nexec %s/%s \"$@\"\n", venvBinDir, binName)

		if runtime.GOOS == "windows" {

			wrapperPath += ".bat"

			wrapperContent = fmt.Sprintf("@echo off\n\"%%~dp0\\%s-app\\venv\\Scripts\\%s\" %%*\n", m.Name, binName)

		}

		if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {

			return fmt.Errorf("failed to create wrapper script: %w", err)

		}

		fmt.Printf("Created wrapper at %s\n", wrapperPath)

		return nil

	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {

		return fmt.Errorf("failed to create target directory: %w", err)
	}

	dstBin := filepath.Join(targetDir, filepath.Base(plan.Bin))
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

func (i *Ingestor) downloadAndUnzip(ctx context.Context, url, dest string) error {
	fmt.Printf("Downloading %s...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "anyisland-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}

	fmt.Printf("Extracting to %s...\n", dest)
	r, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	// Extract files
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	// If the zip contained exactly one directory, move its contents up
	entries, _ := os.ReadDir(dest)
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(dest, entries[0].Name())
		fmt.Printf("Flattening directory structure from %s\n", subDir)
		subEntries, _ := os.ReadDir(subDir)
		for _, se := range subEntries {
			os.Rename(filepath.Join(subDir, se.Name()), filepath.Join(dest, se.Name()))
		}
		os.Remove(subDir)
	}

	return nil
}

func (i *Ingestor) Ingest(ctx context.Context, repoURL string) (*Manifest, error) {
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
				return m, nil
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

		// 1. Detect default branch via GitHub API
		ghRepo, _, err := i.gh.Repositories.Get(ctx, owner, repo)
		defaultBranch := "main"
		if err == nil {
			defaultBranch = ghRepo.GetDefaultBranch()
		}

		// 2. Try to fetch anyisland.json via raw.githubusercontent.com (using curl)
		rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/anyisland.json", owner, repo, defaultBranch)
		fmt.Printf("Attempting to fetch manifest from %s\n", rawURL)

		curlCmd := exec.CommandContext(ctx, "curl", "-fsSL", rawURL)
		output, err := curlCmd.Output()
		if err == nil {
			var m Manifest
			if err := json.Unmarshal(output, &m); err == nil {
				fmt.Println("Found anyisland.json via raw GitHub, using provided build plan.")
				return &m, nil
			}
		}

		// Fetch file tree (simplified)
		tree, _, err := i.gh.Git.GetTree(ctx, owner, repo, defaultBranch, true)
		if err != nil {
			return nil, err
		}

		for _, entry := range tree.Entries {
			files = append(files, entry.GetPath())
		}

		// Fetch README via curl if possible, otherwise API
		readmeURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/README.md", owner, repo, defaultBranch)
		curlReadme := exec.CommandContext(ctx, "curl", "-fsSL", readmeURL)
		readmeOutput, err := curlReadme.Output()
		if err == nil {
			readmeContent = string(readmeOutput)
		} else {
			readme, _, err := i.gh.Repositories.GetReadme(ctx, owner, repo, nil)
			if err == nil {
				content, _ := readme.GetContent()
				readmeContent = content
			}
		}
	}

	// First-class Go support: Detect go.mod and goreleaser
	isGo := false
	hasGoReleaser := false
	for _, f := range files {
		if f == "go.mod" {
			isGo = true
		}
		if strings.Contains(f, ".goreleaser.yaml") || strings.Contains(f, ".goreleaser.yml") {
			hasGoReleaser = true
		}
	}

	if isGo {
		fmt.Println("detected Go project, optimizing build plan...")
		var steps []string
		binName := repo

		if hasGoReleaser {
			fmt.Println("found GoReleaser config, using goreleaser for build...")
			steps = []string{"goreleaser build --snapshot --single-target --clean"}
			// We assume binName stays same or we try to find it in dist/
			binName = "dist/" + repo + "_*/" + repo
			// Note: Real implementation would parse .goreleaser.yaml to find the exact bin
		} else {
			steps = []string{"go build -v -o " + binName}
		}

		return &Manifest{
			Name:    repo,
			Version: "latest",
			Build: agent.BuildPlan{
				Toolchain: "go",
				Steps:     steps,
				Bin:       binName,
			},
		}, nil
	}

	// First-class Flutter support
	isFlutter := false
	for _, f := range files {
		if f == "pubspec.yaml" {
			// We could read the file to be 100% sure, but pubspec.yaml is a very strong indicator
			isFlutter = true
			break
		}
	}

	if isFlutter {
		fmt.Println("detected Flutter project, optimizing build plan...")
		targetOS := runtime.GOOS
		var buildCmd string
		var binPath string

		switch targetOS {
		case "linux":
			buildCmd = "flutter build linux --release"
			binPath = "build/linux/x64/release/bundle/"
		case "darwin":
			buildCmd = "flutter build macos --release"
			binPath = "build/macos/Build/Products/Release/" // Simplified, usually .app
		case "windows":
			buildCmd = "flutter build windows --release"
			binPath = "build/windows/runner/Release/"
		default:
			return nil, fmt.Errorf("flutter build not supported on %s", targetOS)
		}

		return &Manifest{

			Name: repo,

			Version: "latest",

			Build: agent.BuildPlan{

				Toolchain: "flutter",

				Steps: []string{

					"flutter pub get",

					buildCmd,
				},

				Bin: binPath, // For Flutter, we treat the whole bundle as the "binary"

			},
		}, nil

	}

	// First-class Node.js/TS support

	isNode := false

	isTS := false

	for _, f := range files {

		if f == "package.json" {

			isNode = true

		}

		if f == "tsconfig.json" {

			isTS = true

		}

	}

	if isNode {

		fmt.Println("detected Node.js project, optimizing build plan...")

		steps := []string{"npm install"}

		if isTS {

			fmt.Println("detected TypeScript, adding build step...")

			steps = append(steps, "npm run build")

		} else {

			// Check for build script in package.json (simplified heuristic)

			steps = append(steps, "npm run build --if-present")

		}

		return &Manifest{

			Name: repo,

			Version: "latest",

			Build: agent.BuildPlan{

				Toolchain: "node",

				Steps: steps,

				Bin: ".", // For Node, we move the whole directory and find the bin in package.json

			},
		}, nil

	}

	// First-class Rust support

	isRust := false

	for _, f := range files {

		if f == "Cargo.toml" {

			isRust = true

			break

		}

	}

	if isRust {

		fmt.Println("detected Rust project, optimizing build plan...")

		return &Manifest{

			Name: repo,

			Version: "latest",

			Build: agent.BuildPlan{

				Toolchain: "rust",

				Steps: []string{"cargo build --release"},

				Bin: "target/release/" + repo,
			},
		}, nil

	}

	// First-class Python support

	isPython := false

	for _, f := range files {

		if f == "requirements.txt" || f == "pyproject.toml" || f == "setup.py" {

			isPython = true

			break

		}

	}

	if isPython {

		fmt.Println("detected Python project, optimizing build plan...")

		return &Manifest{

			Name: repo,

			Version: "latest",

			Build: agent.BuildPlan{

				Toolchain: "python",

				Steps: []string{

					"python3 -m venv venv",

					"./venv/bin/pip install --upgrade pip",

					"./venv/bin/pip install .", // Try to install as a package

				},

				Bin: "venv", // We mark venv as the target for specialized handling

			},
		}, nil

	}

	// ... (logic to get files and readmeContent) ...

	// Discretion Check
	var discretion agent.DiscretionResult
	if v, ok := i.agent.(*agent.VibeauraSynthesizer); ok {
		fmt.Println("Performing AI-powered discretion check...")
		res, err := v.AnalyzeDiscretion(ctx, files, readmeContent)
		if err == nil {
			discretion = *res
		} else {
			fmt.Printf("AI discretion check failed (%v), falling back to heuristics...\n", err)
			discretion = agent.AnalyzeDiscretion(files, readmeContent)
		}
	} else {
		discretion = agent.AnalyzeDiscretion(files, readmeContent)
	}

	if !discretion.Allowed {
		return nil, fmt.Errorf("Ingestion declined: %s", discretion.Reason)
	}

	fmt.Println("Generating build plan via AI...")

	plan, err := i.agent.GenerateBuildPlan(ctx, repoURL, files, readmeContent)
	if err != nil {
		fmt.Printf("AI build plan generation failed (%v), falling back to basic heuristics...\n", err)
		fallback := &agent.HeuristicSynthesizer{}
		plan, err = fallback.GenerateBuildPlan(ctx, repoURL, files, readmeContent)
		if err != nil {
			return nil, err
		}
	}

	return &Manifest{
		Name:    repo,
		Version: "latest",
		Build:   *plan,
	}, nil
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
