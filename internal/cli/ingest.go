package cli

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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

func (i *Ingestor) Build(ctx context.Context, m *Manifest, repoURL string) (string, string, error) {
	repoURL = normalizeRepoURL(repoURL)
	workDir := i.getSourcePath(repoURL, m.Name)

	// Security: Prevent anyisland from being cloned into its own source dir 
	// (e.g. if someone tries to install anyisland into anyisland)
	if m.Name == "anyisland" {
		exePath, _ := os.Executable()
		absWorkDir, _ := filepath.Abs(workDir)
		if strings.HasPrefix(exePath, absWorkDir) {
			return "", "", fmt.Errorf("refusing to build anyisland within its own source directory (%s)", workDir)
		}
	}

	if _, err := os.Stat(repoURL); err == nil {
		fmt.Printf("Using local source at %s\n", workDir)
	} else if strings.HasSuffix(repoURL, ".zip") || strings.Contains(repoURL, "/zipball/") || strings.Contains(repoURL, "/archive/") {
		if err := i.downloadAndUnzip(ctx, repoURL, workDir); err != nil {
			return "", "", err
		}
	} else {
		if _, err := os.Stat(filepath.Join(workDir, ".git")); os.IsNotExist(err) {
			fmt.Printf("Cloning %s to %s...\n", repoURL, workDir)
			if err := os.MkdirAll(filepath.Dir(workDir), 0755); err != nil {
				return "", "", err
			}
			cmd := exec.CommandContext(ctx, "git", "clone", repoURL, workDir)
			if err := cmd.Run(); err != nil {
				return "", "", fmt.Errorf("git clone failed: %w", err)
			}
		} else {
			fmt.Printf("Updating %s in %s...\n", repoURL, workDir)
			// Ensure we are on a clean state
			_ = exec.CommandContext(ctx, "git", "-C", workDir, "reset", "--hard").Run()
			_ = exec.CommandContext(ctx, "git", "-C", workDir, "clean", "-fd").Run()

			fetchCmd := exec.CommandContext(ctx, "git", "-C", workDir, "fetch", "origin")
			if err := fetchCmd.Run(); err != nil {
				return "", "", fmt.Errorf("git fetch failed: %w", err)
			}

			// Detect default branch if not specified
			branch := "master"
			checkMain := exec.CommandContext(ctx, "git", "-C", workDir, "show-ref", "--verify", "refs/remotes/origin/main")
			if err := checkMain.Run(); err == nil {
				branch = "main"
			}

			resetCmd := exec.CommandContext(ctx, "git", "-C", workDir, "reset", "--hard", "origin/"+branch)
			if err := resetCmd.Run(); err != nil {
				return "", "", fmt.Errorf("git reset to origin/%s failed: %w", branch, err)
			}
		}
	}

	manifestPath := filepath.Join(workDir, "anyisland.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Println("Crafting anyisland.json for this tool...")
		data, err := json.MarshalIndent(m, "", "  ")
		if err == nil {
			_ = os.WriteFile(manifestPath, data, 0644)
		}
	}

	plan := m.Build

	if plan.Toolchain == "go" {
		if _, err := exec.LookPath("go"); err != nil {
			return "", "", fmt.Errorf("go toolchain required but not found in PATH")
		}
	}
	if plan.Toolchain == "rust" {
		if _, err := exec.LookPath("cargo"); err != nil {
			return "", "", fmt.Errorf("rust toolchain (cargo) required but not found in PATH")
		}
	}
	if plan.Toolchain == "python" {
		if _, err := exec.LookPath("python3"); err != nil {
			if _, err := exec.LookPath("python"); err != nil {
				return "", "", fmt.Errorf("python3 or python required but not found in PATH")
			}
		}
	}

	for _, step := range plan.Steps {
		step = strings.TrimSpace(step)
		if step == "" || strings.HasPrefix(step, "#") {
			continue
		}
		fmt.Printf("Executing: %s\n", step)
		fullArgs := strings.Fields(step)
		if len(fullArgs) == 0 {
			continue
		}

		var env []string
		args := fullArgs
		for len(args) > 0 && strings.Contains(args[0], "=") && !strings.HasPrefix(args[0], "-") {
			env = append(env, args[0])
			args = args[1:]
		}

		if len(args) == 0 {
			continue
		}

		// Inject ldflags for Anyisland core components
		if (m.Name == "anyisland") && args[0] == "go" && len(args) > 1 && (args[1] == "build" || args[1] == "install") {
			commit, _ := i.DiscoverLatestCommit(ctx, repoURL)
			if commit == "" {
				commit = "unknown"
			}
			buildTime := time.Now().UTC().Format(time.RFC3339)
			ldflags := fmt.Sprintf("-X github.com/nathfavour/anyisland/internal/cli.Version=%s -X github.com/nathfavour/anyisland/internal/cli.Commit=%s -X github.com/nathfavour/anyisland/internal/cli.BuildTime=%s", m.Version, commit, buildTime)

			// Insert ldflags right after 'build' or 'install'
			newArgs := []string{args[0], args[1], "-ldflags", ldflags}
			newArgs = append(newArgs, args[2:]...)
			args = newArgs
		}

		buildCmd := exec.CommandContext(ctx, args[0], args[1:]...)
		buildCmd.Dir = workDir
		if len(env) > 0 {
			buildCmd.Env = append(os.Environ(), env...)
		}

		var combinedOutput strings.Builder
		buildCmd.Stdout = io.MultiWriter(os.Stdout, &combinedOutput)
		buildCmd.Stderr = io.MultiWriter(os.Stderr, &combinedOutput)

		if err := buildCmd.Run(); err != nil {
			if v, ok := i.agent.(*agent.VibeauraSynthesizer); ok {
				analysis, diagErr := v.DebugBuildFailure(ctx, combinedOutput.String(), m)
				if diagErr == nil {
					fmt.Printf("\nAI Analysis:\n%s\n", analysis)
				}
			}
			return "", "", fmt.Errorf("build step failed: %w", err)
		}
	}

	if plan.Bin == "" {
		return "", "", fmt.Errorf("build plan does not specify a binary ('bin') to install")
	}

	binPattern := filepath.Join(workDir, plan.Bin)
	matches, err := filepath.Glob(binPattern)
	srcBin := ""
	if err == nil && len(matches) > 0 {
		srcBin = matches[0]
	} else {
		srcBin = binPattern
	}

	targetDir := ""
	if m.InstallDir != "" {
		targetDir = m.InstallDir
	} else if plan.InstallDir != "" {
		targetDir = plan.InstallDir
	} else {
		targetDir = i.sys.GetIslandBinDir()
		if plan.Bin == "anyisland" {
		                        targetDir = i.sys.GetBinDir()
		                }	}

	if plan.Toolchain == "flutter" {
		appDir := filepath.Join(targetDir, m.Name+"-app")
		_ = os.RemoveAll(appDir)
		os.MkdirAll(appDir, 0755)
		cpCmd := exec.Command("cp", "-r", srcBin+"/.", appDir)
		if runtime.GOOS == "windows" {
			cpCmd = exec.Command("xcopy", "/E", "/I", "/Y", srcBin, appDir)
		}
		cpCmd.Run()
		entries, _ := os.ReadDir(appDir)
		exeName := m.Name
		for _, entry := range entries {
			if !entry.IsDir() && !strings.Contains(entry.Name(), ".") {
				exeName = entry.Name()
				break
			}
		}
		wrapperPath := filepath.Join(targetDir, m.Name)
		wrapperContent := fmt.Sprintf("#!/bin/bash\ncd %s && ./%s \"$@\"\n", appDir, exeName)
		os.WriteFile(wrapperPath, []byte(wrapperContent), 0755)
		return "", wrapperPath, nil
	}

	if plan.Toolchain == "node" {
		appDir := filepath.Join(targetDir, m.Name+"-app")
		_ = os.RemoveAll(appDir)
		os.MkdirAll(appDir, 0755)
		cpCmd := exec.Command("cp", "-r", workDir+"/.", appDir)
		cpCmd.Run()
		pkgData, _ := os.ReadFile(filepath.Join(appDir, "package.json"))
		var binScript string
		var pkg struct {
			Bin map[string]string `json:"bin"`
		}
		json.Unmarshal(pkgData, &pkg)
		for _, script := range pkg.Bin {
			binScript = script
			break
		}
		if binScript == "" {
			binScript = "index.js"
		}
		wrapperPath := filepath.Join(targetDir, m.Name)
		wrapperContent := fmt.Sprintf("#!/bin/bash\nnode %s/%s \"$@\"\n", appDir, binScript)
		os.WriteFile(wrapperPath, []byte(wrapperContent), 0755)
		return "", wrapperPath, nil
	}

	if plan.Toolchain == "python" {
		appDir := filepath.Join(targetDir, m.Name+"-app")
		_ = os.RemoveAll(appDir)
		os.MkdirAll(appDir, 0755)
		cpCmd := exec.Command("cp", "-r", workDir+"/.", appDir)
		cpCmd.Run()
		venvBinDir := filepath.Join(appDir, "venv", "bin")
		wrapperPath := filepath.Join(targetDir, m.Name)
		wrapperContent := fmt.Sprintf("#!/bin/bash\nexec %s/%s \"$@\"\n", venvBinDir, m.Name)
		os.WriteFile(wrapperPath, []byte(wrapperContent), 0755)
		return "", wrapperPath, nil
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create target directory: %w", err)
	}

	dstBin := filepath.Join(targetDir, filepath.Base(plan.Bin))
	fmt.Printf("Installing %s to %s...\n", srcBin, dstBin)

	if _, err := os.Stat(dstBin); err == nil {
		oldBin := dstBin + ".bak"
		_ = os.Remove(oldBin)
		os.Rename(dstBin, oldBin)
	}

	input, err := os.ReadFile(srcBin)
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(dstBin, input, 0755); err != nil {
		return "", "", err
	}

	hash := calculateFileHash(dstBin)
	return hash, dstBin, nil
}

func (i *Ingestor) VerifyToolIntegrity(path string, expectedHash string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	if expectedHash == "" {
		return true
	}
	actual := calculateFileHash(path)
	return actual == expectedHash
}

func calculateFileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (i *Ingestor) downloadAndUnzip(ctx context.Context, url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	tmpFile, _ := os.CreateTemp("", "anyisland-*.zip")
	defer os.Remove(tmpFile.Name())
	io.Copy(tmpFile, resp.Body)
	r, _ := zip.OpenReader(tmpFile.Name())
	defer r.Close()
	os.MkdirAll(dest, 0755)
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		outFile, _ := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		rc, _ := f.Open()
		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}

func (i *Ingestor) Ingest(ctx context.Context, repoURL string) (*Manifest, string, error) {
	repoURL = normalizeRepoURL(repoURL)
	commit, _ := i.DiscoverLatestCommit(ctx, repoURL)

	var owner, repo string
	var files []string
	var readmeContent string

	if _, err := os.Stat(repoURL); err == nil {
		absPath, _ := filepath.Abs(repoURL)
		manifestPath := filepath.Join(absPath, "anyisland.json")
		if _, err := os.Stat(manifestPath); err == nil {
			m, err := LoadManifest(manifestPath)
			if err == nil {
				return m, commit, nil
			}
		}

		repo = filepath.Base(absPath)
		filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || (info.IsDir() && info.Name() == ".git") {
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
	} else {
		parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
		if len(parts) < 2 {
			return nil, commit, fmt.Errorf("invalid GitHub URL: %s", repoURL)
		}
		owner = parts[0]
		repo = parts[1]

		ghRepo, _, err := i.gh.Repositories.Get(ctx, owner, repo)
		defaultBranch := "main"
		if err == nil && ghRepo != nil {
			defaultBranch = ghRepo.GetDefaultBranch()
		}

		rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/anyisland.json", owner, repo, defaultBranch)
		curlCmd := exec.CommandContext(ctx, "curl", "-fsSL", rawURL)
		output, err := curlCmd.Output()
		if err == nil {
			var m Manifest
			if err := json.Unmarshal(output, &m); err == nil {
				return &m, commit, nil
			}
		}

		tree, _, err := i.gh.Git.GetTree(ctx, owner, repo, defaultBranch, true)
		if err == nil && tree != nil {
			for _, entry := range tree.Entries {
				files = append(files, entry.GetPath())
			}
		}

		readme, _, err := i.gh.Repositories.GetReadme(ctx, owner, repo, nil)
		if err == nil && readme != nil {
			content, _ := readme.GetContent()
			readmeContent = content
		}
	}

		// AI Discretion Check

		discretion, err := i.agent.AnalyzeDiscretion(ctx, files, readmeContent)

		if err == nil && discretion != nil && !discretion.Allowed {

			return nil, commit, fmt.Errorf("repository is not a buildable tool: %s", discretion.Reason)

		}

	

		manifest := &Manifest{

			Name:    repo,

			Version: "latest",

		}

	

		isGo := false

		for _, f := range files {

			if f == "go.mod" {

				isGo = true

				break

			}

		}

		if isGo {

			manifest.Build = agent.BuildPlan{

				Toolchain: "go",

				Steps:     []string{"go build -v -o " + repo},

				Bin:       repo,

			}

		} else {

			isRust := false

			for _, f := range files {

				if f == "Cargo.toml" {

					isRust = true

					break

				}

			}

			if isRust {

				manifest.Build = agent.BuildPlan{

					Toolchain: "rust",

					Steps:     []string{"cargo build --release"},

					Bin:       "target/release/" + repo,

				}

			} else {

				plan, err := i.agent.GenerateBuildPlan(ctx, repoURL, files, readmeContent)

				if err != nil {

					fallback := &agent.HeuristicSynthesizer{}

					plan, _ = fallback.GenerateBuildPlan(ctx, repoURL, files, readmeContent)

				}

				manifest.Build = *plan

			}

		}

	

		// Self-Conflict Prevention: If the tool identifies as anyisland, 
		// we must ensure it doesn't overwrite the manager via a generic install.

		return manifest, commit, nil
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
