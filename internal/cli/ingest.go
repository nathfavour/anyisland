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
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/deps"
	"github.com/nathfavour/anyisland/internal/pal"
)

type Ingestor struct {
	gh       *github.Client
	agent    agent.Synthesizer
	sys      pal.System
	resolver *Resolver
	cfg      *Config
}

func NewIngestor(ag agent.Synthesizer, sys pal.System, cfg *Config) *Ingestor {
	if cfg == nil {
		cm := NewConfigManager(sys)
		cfg, _ = cm.Load()
	}
	return &Ingestor{
		gh:       github.NewClient(nil),
		agent:    ag,
		sys:      sys,
		resolver: NewResolver(),
		cfg:      cfg,
	}
}

func (i *Ingestor) ParseVersion(input string) (string, string) {
	if strings.Contains(input, "@") {
		parts := strings.Split(input, "@")
		return parts[0], parts[1]
	}
	return input, ""
}

func (i *Ingestor) DiscoverRelease(ctx context.Context, owner, repo, version string) (*github.RepositoryRelease, error) {
	if version == "" || version == "latest" {
		release, _, err := i.gh.Repositories.GetLatestRelease(ctx, owner, repo)
		return release, err
	}

	// Try as a tag
	release, _, err := i.gh.Repositories.GetReleaseByTag(ctx, owner, repo, version)
	if err == nil {
		return release, nil
	}

	// If not a specific release tag, list and find
	releases, _, err := i.gh.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}

	for _, rel := range releases {
		if rel.GetTagName() == version || rel.GetName() == version {
			return rel, nil
		}
	}

	return nil, fmt.Errorf("release %s not found", version)
}

func (i *Ingestor) MatchAsset(ctx context.Context, assets []*github.ReleaseAsset) *github.ReleaseAsset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	type candidate struct {
		asset *github.ReleaseAsset
		score int
	}

	var candidates []candidate

	for _, asset := range assets {
		name := strings.ToLower(asset.GetName())
		score := 0

		// OS matching
		if strings.Contains(name, goos) {
			score += 10
		} else if goos == "darwin" && (strings.Contains(name, "macos") || strings.Contains(name, "apple") || strings.Contains(name, "osx")) {
			score += 10
		} else if goos == "linux" && (strings.Contains(name, "linux") || strings.Contains(name, "musl") || strings.Contains(name, "gnu")) {
			score += 10
		}

		// Architecture matching
		if strings.Contains(name, goarch) {
			score += 5
		} else if goarch == "amd64" && (strings.Contains(name, "x86_64") || strings.Contains(name, "x64") || strings.Contains(name, "intel")) {
			score += 5
		} else if goarch == "arm64" && (strings.Contains(name, "aarch64") || strings.Contains(name, "armv8")) {
			score += 5
		}

		// Penalize debug or source assets if we want binaries
		if strings.Contains(name, "src") || strings.Contains(name, "source") || strings.Contains(name, "dev") || strings.Contains(name, "dbg") {
			score -= 20
		}

		// Prefer common archive formats or raw binaries
		if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tgz") || strings.HasSuffix(name, ".xz") {
			score += 2
		}

		if score > 0 {
			candidates = append(candidates, candidate{asset: asset, score: score})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by score
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// If the top score is high and clear winner, take it
	if candidates[0].score >= 15 {
		if len(candidates) == 1 || candidates[0].score > candidates[1].score+2 {
			return candidates[0].asset
		}
	}

	// Ambiguity or low confidence -> Use AI discretion
	fmt.Println("🤖 Using AI to select the best binary for your system...")
	var assetNames []string
	nameToAsset := make(map[string]*github.ReleaseAsset)
	for _, c := range candidates {
		assetNames = append(assetNames, c.asset.GetName())
		nameToAsset[c.asset.GetName()] = c.asset
	}

	selectedName, err := i.agent.SelectAsset(ctx, assetNames, goos, goarch)
	if err == nil && selectedName != "" {
		if asset, ok := nameToAsset[selectedName]; ok {
			return asset
		}
	}

	// Fallback to top heuristic if AI fails
	if candidates[0].score >= 10 {
		return candidates[0].asset
	}

	return nil
}

func (i *Ingestor) DiscoverLatestCommit(ctx context.Context, repoURL string) (string, error) {
	repoURL = normalizeRepoURL(repoURL)

	// Optimization: If it's a local path, try to get the local HEAD directly
	if _, err := os.Stat(repoURL); err == nil {
		cmd := exec.CommandContext(ctx, "git", "-C", repoURL, "rev-parse", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}

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

func (i *Ingestor) GetSourcePath(repoURL string, pkgName string) string {
	repoURL = normalizeRepoURL(repoURL)
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

type ResolvedPackage struct {
	Manifest *Manifest
	Source   string
	Commit   string
}

func (i *Ingestor) ResolveDependencies(ctx context.Context, rootSource string) ([]ResolvedPackage, error) {
	graph := deps.NewGraph()
	manifests := make(map[string]*Manifest)
	sources := make(map[string]string)
	commits := make(map[string]string)

	var resolve func(source string) error
	resolve = func(source string) error {
		m, commit, finalURL, err := i.Ingest(ctx, source)
		if err != nil {
			return fmt.Errorf("failed to ingest %s: %w", source, err)
		}

		if _, ok := manifests[m.Name]; ok {
			return nil
		}

		manifests[m.Name] = m
		sources[m.Name] = finalURL
		commits[m.Name] = commit

		for _, dep := range m.Dependencies {
			graph.AddDependency(m.Name, dep)
			if err := resolve(dep); err != nil {
				return err
			}
		}
		return nil
	}

	if err := resolve(rootSource); err != nil {
		return nil, err
	}

	rootM, _, _, _ := i.Ingest(ctx, rootSource)

	resolvedNames, err := graph.Resolve(rootM.Name, func(name string) ([]string, error) {
		if m, ok := manifests[name]; ok {
			return m.Dependencies, nil
		}
		return nil, fmt.Errorf("manifest not found for %s", name)
	})
	if err != nil {
		return nil, err
	}

	var result []ResolvedPackage
	for _, name := range resolvedNames {
		result = append(result, ResolvedPackage{
			Manifest: manifests[name],
			Source:   sources[name],
			Commit:   commits[name],
		})
	}

	return result, nil
}

func (i *Ingestor) Build(ctx context.Context, m *Manifest, repoURL string) (string, string, error) {
	repoURL = normalizeRepoURL(repoURL)
	workDir := i.GetSourcePath(repoURL, m.Name)

	if m.SourceDir != "" {
		customDir := i.ExpandPath(m.SourceDir)
		if filepath.IsAbs(customDir) {
			workDir = customDir
		} else {
			// Relative SourceDir in manifest should be relative to the repo root
			workDir = filepath.Join(workDir, customDir)
		}
	}

	// Binary Release Path

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
			args := i.cloneArgs(repoURL, workDir, m)
			cmd := exec.CommandContext(ctx, "git", args...)
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

			// Detect branch: use config, then version, then default
			branch := i.cfg.Install.DefaultBranch
			if branch == "" {
				branch = "master"
				checkMain := exec.CommandContext(ctx, "git", "-C", workDir, "show-ref", "--verify", "refs/remotes/origin/main")
				if err := checkMain.Run(); err == nil {
					branch = "master"
				}
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
                if plan.Toolchain == "shell" {
                        buildCmd := exec.CommandContext(ctx, "bash", "-c", step)
                        buildCmd.Dir = workDir
                        var combinedOutput strings.Builder
                        buildCmd.Stdout = io.MultiWriter(os.Stdout, &combinedOutput)
                        buildCmd.Stderr = io.MultiWriter(os.Stderr, &combinedOutput)
                        if err := buildCmd.Run(); err != nil {
                                return "", "", fmt.Errorf("shell step failed: %w", err)
                        }
                        continue
                }
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
		}
	}
	targetDir = i.ExpandPath(targetDir)

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

		// Clean up the build output in source directory
		_ = os.RemoveAll(srcBin)

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

	binName := filepath.Base(plan.Bin)
	if m.BinName != "" {
		binName = m.BinName
	}
	dstBin := filepath.Join(targetDir, binName)
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

	// Create aliases
	for _, alias := range m.Aliases {
		if alias == binName {
			continue
		}
		aliasPath := filepath.Join(targetDir, alias)
		_ = os.Remove(aliasPath)
		fmt.Printf("Creating alias: %s -> %s\n", alias, binName)
		_ = os.Symlink(binName, aliasPath)
	}

	// Clean up: Remove the original binary from the source directory to prevent bloat
	_ = os.Remove(srcBin)

	if err := i.buildSubmoduleTargets(ctx, workDir, m); err != nil {
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

func (i *Ingestor) Ingest(ctx context.Context, repoURL string) (*Manifest, string, string, error) {
	originalURL := repoURL
	repoURL, version := i.ParseVersion(normalizeRepoURL(repoURL))

	// If it's a simple name, try to discover it
	if !strings.Contains(originalURL, "/") && !strings.Contains(originalURL, ".") && !strings.Contains(originalURL, ":") {
		// 1. Check if it's an official package in the current source tree
		officialPath := filepath.Join("packages", "official", repoURL, "anyisland.json")
		if _, err := os.Stat(officialPath); err == nil {
			m, err := LoadManifest(officialPath)
			if err == nil && m.Repository != "" {
				repoURL = normalizeRepoURL(m.Repository)
				fmt.Printf("Using official package: %s -> %s\n", officialPath, repoURL)
			} else {
				repoURL, _ = filepath.Abs(filepath.Dir(officialPath))
				fmt.Printf("Using official package path: %s\n", officialPath)
			}
		} else {
			discovered, err := i.resolver.Resolve(ctx, repoURL)
			if err == nil && discovered != "" {
				repoURL = normalizeRepoURL(discovered)
			} else {
				// Final fallback to AI if resolver fails
				fmt.Printf("Deep searching for tool: %s...\n", repoURL)
				discovered, err := i.agent.DiscoverTool(ctx, repoURL)
				if err == nil && discovered != "" && discovered != "NONE" {
					repoURL = normalizeRepoURL(discovered)
				} else {
					return nil, "", "", fmt.Errorf("could not find a repository for '%s'", repoURL)
				}
			}
		}
	}

	commit, _ := i.DiscoverLatestCommit(ctx, repoURL)
	finalURL := repoURL

	var owner, repo string
	var files []string
	var readmeContent string
	var manifest *Manifest

	if _, err := os.Stat(repoURL); err == nil {
		absPath, _ := filepath.Abs(repoURL)
		manifestPath := filepath.Join(absPath, "anyisland.json")
		if _, err := os.Stat(manifestPath); err == nil {
			m, err := LoadManifest(manifestPath)
			if err == nil {
				manifest = m
			}
		}

		repo = filepath.Base(absPath)
		if manifest != nil && manifest.Name != "" {
			repo = manifest.Name
		}

		if manifest == nil {
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
		}
	} else {
		parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
		if len(parts) < 2 {
			return nil, commit, finalURL, fmt.Errorf("invalid GitHub URL: %s", repoURL)
		}
		owner = parts[0]
		repo = parts[1]

                // Smarter Branch Discovery
                targetRef := version
                if targetRef == "" {
                        targetRef = i.cfg.Install.DefaultBranch
                }
                if targetRef == "" {
                        targetRef = i.DiscoverDefaultBranch(ctx, repoURL)
                }

                // Try to load anyisland.json first
                rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/anyisland.json", owner, repo, targetRef)
		curlCmd := exec.CommandContext(ctx, "curl", "-fsSL", rawURL)
		output, err := curlCmd.Output()
		if err == nil {
			var m Manifest
			if err := json.Unmarshal(output, &m); err == nil {
				manifest = &m
			}
		}

		// Smart Discovery for Binary Release
		if i.cfg.Install.Preference != "source" {
			release, err := i.DiscoverRelease(ctx, owner, repo, version)
			if err == nil && release != nil {
				asset := i.MatchAsset(ctx, release.Assets)
				if asset != nil {
					fmt.Printf("✨ Found binary release: %s (%s)\n", release.GetTagName(), asset.GetName())
					if manifest == nil {
						manifest = &Manifest{Name: repo}
					}
					manifest.Version = release.GetTagName()
					manifest.Release = &ReleaseInfo{
						TagName:     release.GetTagName(),
						AssetURL:    asset.GetBrowserDownloadURL(),
						AssetName:   asset.GetName(),
						IsBinary:    true,
						PublishedAt: release.GetPublishedAt().String(),
					}
					return manifest, commit, finalURL, nil
				}
			}
		}

		if manifest == nil {
			tree, _, err := i.gh.Git.GetTree(ctx, owner, repo, targetRef, true)
			if err == nil && tree != nil {
				for _, entry := range tree.Entries {
					files = append(files, entry.GetPath())
				}
			}

			readme, _, err := i.gh.Repositories.GetReadme(ctx, owner, repo, &github.RepositoryContentGetOptions{Ref: targetRef})
			if err == nil && readme != nil {
				content, _ := readme.GetContent()
				readmeContent = content
			}
		}
	}

	if manifest != nil {
		if version != "" {
			manifest.Version = version
		}
		return manifest, commit, finalURL, nil
	}

	// AI Discretion Check
	discretion, err := i.agent.AnalyzeDiscretion(ctx, files, readmeContent)
	if err == nil && discretion != nil && !discretion.Allowed {
		return nil, commit, finalURL, fmt.Errorf("repository is not a buildable tool: %s", discretion.Reason)
	}

	// Use AI to generate a build plan if no manifest exists
	aiPlan, err := i.agent.GenerateBuildPlan(ctx, repoURL, files, readmeContent)
	if err == nil && aiPlan != nil {
		name := repo
		if aiPlan.Name != "" {
			name = aiPlan.Name
		}
		manifest = &Manifest{
			Name:    name,
			Version: "latest",
			Build:   *aiPlan,
		}
		return manifest, commit, finalURL, nil
	}

	manifest = &Manifest{
		Name:    repo,
		Version: "latest",
	}

	isGo := false
	isAnyisland := false
	for _, f := range files {
		if f == "go.mod" {
			isGo = true
		}
		if f == "cmd/anyisland/main.go" {
			isAnyisland = true
		}
	}

	if isGo {
		if isAnyisland {
			manifest.Build = agent.BuildPlan{
				Name:      "anyisland",
				Toolchain: "go",
				Steps:     []string{"go build -v -o anyisland ./cmd/anyisland"},
				Bin:       "anyisland",
			}
		} else {
			manifest.Build = agent.BuildPlan{
				Name:      repo,
				Toolchain: "go",
				Steps:     []string{"go build -v -o " + repo},
				Bin:       repo,
			}
		}
		manifest.Name = manifest.Build.Name
	}

	return manifest, commit, finalURL, nil
}


func (i *Ingestor) DiscoverDefaultBranch(ctx context.Context, repoURL string) string {
	// 1. Try Git discovery (works for any provider, no API token needed)
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--symref", repoURL, "HEAD")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ref: refs/heads/") {
				parts := strings.Fields(line)
				return strings.TrimPrefix(parts[1], "refs/heads/")
			}
		}
	}

	// 2. Fallback to common branch names
	return "main"
}

func normalizeRepoURL(url string) string {
	url = ExpandPath(url)
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "git@") {
		// Only check if it's a local path if it starts with . or /
		if strings.HasPrefix(url, ".") || strings.HasPrefix(url, "/") {
			if _, err := os.Stat(url); err == nil {
				abs, _ := filepath.Abs(url)
				return abs
			}
		}

		// If it's a simple name (no dots, no slashes), don't prefix with https:// yet
		// because Ingest will try to discover it.
		if !strings.Contains(url, ".") && !strings.Contains(url, "/") {
			return url
		}

		if strings.Count(url, "/") == 1 {
			return "https://github.com/" + url
		}
		return "https://" + url
	}
	return url
}

func (i *Ingestor) ExpandPath(path string) string {
	return ExpandPath(path)
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	return path
}
