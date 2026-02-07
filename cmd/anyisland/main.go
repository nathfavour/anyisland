package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/cli"
	"github.com/nathfavour/anyisland/internal/crypto"
	"github.com/nathfavour/anyisland/internal/history"
	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/nathfavour/anyisland/internal/registry"
	"github.com/nathfavour/anyisland/pkg/discovery"
		"github.com/spf13/cobra"
	)
	
	var (
		sourceFlag string
	rootCmd = &cobra.Command{
		Use:   "anyisland",
		Short: "Anyisland is an AI-powered package manager",
		Long:  `Anyisland is an AI-powered, platform-agnostic, and decentralized package manager.`,
		                                PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		                                        // Skip check for setup and init command
		                                        if cmd.Name() == "setup" || cmd.Name() == "init" || cmd.Name() == "version" {
		                                                return nil
		                                        }
		                                        sys, err := pal.New()
		                                        if err != nil {
		                                                return nil // Ignore PAL errors in PreRun to allow emergency use
		                                        }
		                
		                                        // Ensure Config Exists
		                                        cm := cli.NewConfigManager(sys)
		                                        cfg, err := cm.Load()
		                                        if err == nil {
		                                                // Save it back to ensure the file exists with defaults if it was new
		                                                cm.Save(cfg)
		                                        }
		                
		                                                                                                // Seamless Auto-Update
		                                                                                                ag := getSynthesizer()
		                                                                                                lm := cli.NewLifecycleManager(sys)
		                                                                                                // Skip auto-update if we are already running the update command
		                                                                                                if cmd.Name() != "update" {
		                                                                                                        lm.BackgroundAutoUpdate(cmd.Context(), ag)
		                                                                                                }		                
		                                        return sys.EnsurePath()
		                                },
		                        }
		                
		                        configCmd = &cobra.Command{
		                                Use:   "config",
		                                Short: "Manage Anyisland configurations",
		                        }
		                
		                        configSetCmd = &cobra.Command{
		                                Use:   "set [key] [value]",
		                                Short: "Set a configuration value",
		                                Args:  cobra.ExactArgs(2),
		                                RunE: func(cmd *cobra.Command, args []string) error {
		                                        sys, err := pal.New()
		                                        if err != nil {
		                                                return err
		                                        }
		                                        cm := cli.NewConfigManager(sys)
		                                        cfg, err := cm.Load()
		                                        if err != nil {
		                                                return err
		                                        }
		                
		                                        key := args[0]
		                                        val := args[1]
		                
		                                        switch key {
		                                        case "update.auto_update":
		                                                cfg.Update.AutoUpdate = (val == "true")
		                                        default:
		                                                return fmt.Errorf("unknown config key: %s", key)
		                                        }
		                
		                                        if err := cm.Save(cfg); err != nil {
		                                                return err
		                                        }
		                                        fmt.Printf("Config %s set to %s\n", key, val)
		                                        return nil
		                                },
		                        }
		                
		                        configShowCmd = &cobra.Command{
		                                Use:   "show",
		                                Short: "Show current configuration",
		                                RunE: func(cmd *cobra.Command, args []string) error {
		                                        sys, err := pal.New()
		                                        if err != nil {
		                                                return err
		                                        }
		                                        cm := cli.NewConfigManager(sys)
		                                        cfg, err := cm.Load()
		                                        if err != nil {
		                                                return err
		                                        }
		                
		                                        fmt.Println("Anyisland Configuration:")
		                                        fmt.Printf("  update.auto_update: %v\n", cfg.Update.AutoUpdate)
		                                        return nil
		                                },
		                        }
		                

	historyCmd = &cobra.Command{
		Use:   "history",
		Short: "Manage E2EE shell history",
	}

	historyRecordCmd = &cobra.Command{
		Use:   "record [command]",
		Short: "Record and sync a command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sys, err := pal.New()
			if err != nil {
				return err
			}
			fullCmd := strings.Join(args, " ")
			ag := getSynthesizer()
			mgr := history.NewManager(sys, ag)

			if err := mgr.SyncCommand(cmd.Context(), fullCmd); err != nil {
				return err
			}
			fmt.Println("Command synced securely.")
			return nil
		},
	}

	historyShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show synced history",
		RunE: func(cmd *cobra.Command, args []string) error {
			sys, err := pal.New()
			if err != nil {
				return err
			}
			ag := getSynthesizer()
			mgr := history.NewManager(sys, ag)

			lines, err := mgr.GetHistory()
			if err != nil {
				return err
			}

			for _, line := range lines {
				fmt.Print(line)
			}
			return nil
		},
	}

	        setupCmd = &cobra.Command{

	                Use:   "setup",

	                Short: "Initialize Anyisland environment",

	                RunE: func(cmd *cobra.Command, args []string) error {

	                        sys, err := pal.New()

	                        if err != nil {

	                                return err

	                        }

	                        if err := sys.InitFolders(); err != nil {

	                                return err

	                        }

	

	                        // Check if paths are already in session

	                        binDir := sys.GetBinDir()

	                        islandBinDir := sys.GetIslandBinDir()

	                        

	                        // We use the internal pal logic to check if injection is needed

	                        fmt.Printf("Anyisland needs to add the following to your PATH:\n  - %s\n  - %s\n", binDir, islandBinDir)

	                        fmt.Print("Would you like Anyisland to automate this for you? (y/N): ")

	                        

	                        var response string

	                        fmt.Scanln(&response)

	                        

	                        if strings.ToLower(response) == "y" {

	                                if err := sys.InjectPath(); err != nil {

	                                        fmt.Printf("Warning: failed to inject PATH: %v\n", err)

	                                } else {

	                                        fmt.Println("✅ PATH updated in your shell configuration.")

	                                }

	                        } else {

	                                fmt.Println("Skipping PATH injection. Please add the directories manually to your shell config.")

	                        }

	

	                        // Initialize Master Key

	
			cm := crypto.NewManager(sys)
			_, err = cm.GetEncryptionKey()
			if err != nil {
				fmt.Printf("Warning: failed to initialize encryption key: %v\n", err)
			}

			_, err = registry.Open(sys.GetIslandDir())
			if err != nil {
				return err
			}
			fmt.Printf("Anyisland initialized at %s\n", sys.GetIslandDir())
			fmt.Println("Please restart your shell or source your rc file to update PATH.")
			return nil
		},
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a project with anyisland.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.InitProject("."); err != nil {
				return err
			}
			fmt.Println("anyisland.json created.")
			return nil
		},
	}

	installCmd = &cobra.Command{
		Use:   "install [url]",
		Short: "Install a tool from a GitHub URL or local path",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := ""
			if len(args) > 0 {
				url = args[0]
			}
			if sourceFlag != "" {
				url = sourceFlag
			}
			                        if url == "" {
			                                return fmt.Errorf("must provide a URL, a tool name, or use --source")
			                        }
			                        
			                        // Check if the URL is actually a search query (no dots, no slashes)
			                        if !strings.Contains(url, "/") && !strings.Contains(url, ".") {
			                                fmt.Printf("Searching for tool: %s...\n", url)
			                                ag := getSynthesizer()
			                                discovered, err := ag.DiscoverTool(cmd.Context(), url)
			                                if err != nil || discovered == "" || discovered == "NONE" {
			                                        return fmt.Errorf("could not find a tool matching '%s'. Please provide a full GitHub URL", url)
			                                }
			                                fmt.Printf("Found: %s\n", discovered)
			                                url = discovered
			                        }
			
			                        sys, err := pal.New()
			
			if err != nil {
				return err
			}

			                        ag := getSynthesizer()
			                        ingestor := cli.NewIngestor(ag, sys)
			
			                                                                        manifest, commit, err := ingestor.Ingest(cmd.Context(), url)
			                                                                        if err != nil {
			                                                                                return err
			                                                                        }
			                        
			                                                                        if manifest.Name == "anyisland" {
			                                                                                return fmt.Errorf("this repository identifies as Anyisland. Use 'anyisland update' to manage the manager")
			                                                                        }
			                        
			                                                                                                                                                                        reg, err := registry.Open(sys.GetIslandDir())
			                                                                                                                                                                        if err != nil {
			                                                                                                                                return err
			                                                                                                                        }
							defer reg.Close()
			
			                                                                        existing, _ := reg.GetTool(manifest.Name)
			                                                                        if existing != nil {
			                                                                                if existing.LastCommit == commit {
			                                                                                        if ingestor.VerifyToolIntegrity(existing.InstallPath, existing.BinaryHash) {
			                                                                                                fmt.Printf("%s is already up-to-date and healthy (commit: %s)\n", manifest.Name, cli.ShortCommit(commit))
			                                                                                                return nil
			                                                                                        }
			                                                                                        fmt.Printf("%s seems broken. Reinstalling...\n", manifest.Name)
			                                                                                } else {
			                                                                                        fmt.Printf("Updating %s from %s to %s...\n", manifest.Name, cli.ShortCommit(existing.LastCommit), cli.ShortCommit(commit))
			                                                                                }
			                                                                        }			
			                        fmt.Println("\nProposed Build Plan:")
			                        for _, step := range manifest.Build.Steps {
			                                fmt.Printf("  - %s\n", step)
			                        }
			
			                        fmt.Println("\nExecuting build...")
			                        hash, installPath, err := ingestor.Build(cmd.Context(), manifest, url)
			                        if err != nil {
			                                return err
			                        }
			
			                        err = reg.RegisterTool(registry.Tool{
			                                Name:        manifest.Name,
			                                Source:      url,
			                                Version:     manifest.Version,
			                                LastCommit:  commit,
			                                BinaryHash:  hash,
			                                InstallPath: installPath,
			                                Type:        "source",
			                        })
						
			if err != nil {
				return err
			}

			                        fmt.Printf("\nSuccessfully installed %s!\n", manifest.Build.Bin)

			
			return nil
		},
	}

	        listCmd = &cobra.Command{

	                Use:   "list",

	                Short: "List registered tools",

	                RunE: func(cmd *cobra.Command, args []string) error {

	                        sys, err := pal.New()

	                        if err != nil {

	                                return err

	                        }

	                        reg, err := registry.Open(sys.GetIslandDir())

	                        if err != nil {

	                                return err

	                        }

	                        defer reg.Close()

	

	                        tools, err := reg.ListTools()

	                        if err != nil {

	                                return err

	                        }

	

	                        if len(tools) == 0 {

	                                fmt.Println("No tools registered.")

	                                return nil

	                        }

	

	                        fmt.Println("Registered tools:")

	                        for _, t := range tools {

	                                fmt.Printf("- %s (%s) [%s]\n", t.Name, t.Version, t.Source)

	                        }

	                        return nil

	                },

	        }

	

	                shellCmd = &cobra.Command{

	

	                        // ... (shell command logic) ...

	

	                }

	

	        

	

	                explainCmd = &cobra.Command{

	

	                        Use:   "explain [tool]",

	

	                        Short: "Get an AI-powered explanation of an installed tool",

	

	                        Args:  cobra.ExactArgs(1),

	

	                        RunE: func(cmd *cobra.Command, args []string) error {

	

	                                toolName := args[0]

	

	                                sys, err := pal.New()

	

	                                if err != nil {

	

	                                        return err

	

	                                }

	

	                                reg, err := registry.Open(sys.GetIslandDir())

	

	                                if err != nil {

	

	                                        return err

	

	                                }

	

	                                defer reg.Close()

	

	        

	

	                                tools, err := reg.ListTools()

	

	                                if err != nil {

	

	                                        return err

	

	                                }

	

	        

	

	                                var targetTool *registry.Tool

	

	                                for _, t := range tools {

	

	                                        if t.Name == toolName {

	

	                                                targetTool = &t

	

	                                                break

	

	                                        }

	

	                                }

	

	        

	

	                                if targetTool == nil {

	

	                                        return fmt.Errorf("tool %s not found", toolName)

	

	                                }

	

	        

	

	                                fmt.Printf("Analyzing %s...\n", toolName)

	

	                                

	

	                                // Load manifest and README

	

	                                appDir := filepath.Join(sys.GetIslandBinDir(), toolName+"-app")

	

	                                manifest, _ := cli.LoadManifest(filepath.Join(appDir, "anyisland.json"))

	

	                                readme, _ := os.ReadFile(filepath.Join(appDir, "README.md"))

	

	                                if len(readme) == 0 {

	

	                                        readme, _ = os.ReadFile(filepath.Join(appDir, "readme.md"))

	

	                                }

	

	        

	

	                                ag := getSynthesizer()

	

	                                explanation, err := ag.ExplainTool(cmd.Context(), toolName, manifest, string(readme))

	

	                                if err != nil {

	

	                                        return err

	

	                                }

	

	        

	

	                                fmt.Printf("\n💡 %s\n", explanation)

	

	                                return nil

	

	                        },

	

	                }

	

	        

	

	                ingestCmd = &cobra.Command{

	

	        

	
		Use:   "ingest [url]",
		Short: "Ingest a tool from a GitHub URL",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := ""
			if len(args) > 0 {
				url = args[0]
			}
			if sourceFlag != "" {
				url = sourceFlag
			}
			if url == "" {
				return fmt.Errorf("must provide a URL or use --source")
			}
			sys, err := pal.New()
			if err != nil {
				return err
			}

			                        ag := getSynthesizer()
			                        ingestor := cli.NewIngestor(ag, sys)
			
			                                                                        manifest, commit, err := ingestor.Ingest(cmd.Context(), url)
			                                                                        if err != nil {
			                                                                                return err
			                                                                        }
			                        
			                                                                        if manifest.Name == "anyisland" {
			                                                                                return fmt.Errorf("this repository identifies as Anyisland. Use 'anyisland update' to manage the manager")
			                                                                        }
			                        
			                                                                                                                                                                        reg, err := registry.Open(sys.GetIslandDir())
			                                                                                                                                                                        if err != nil {
			                                                                                                                                return err
			                                                                                                                        }
							defer reg.Close()
			
			                                                                        existing, _ := reg.GetTool(manifest.Name)
			                                                                        if existing != nil {
			                                                                                if existing.LastCommit == commit {
			                                                                                        if ingestor.VerifyToolIntegrity(existing.InstallPath, existing.BinaryHash) {
			                                                                                                fmt.Printf("%s is already up-to-date and healthy (commit: %s)\n", manifest.Name, cli.ShortCommit(commit))
			                                                                                                return nil
			                                                                                        }
			                                                                                        fmt.Printf("%s seems broken. Reinstalling...\n", manifest.Name)
			                                                                                } else {
			                                                                                        fmt.Printf("Updating %s from %s to %s...\n", manifest.Name, cli.ShortCommit(existing.LastCommit), cli.ShortCommit(commit))
			                                                                                }
			                                                                        }			
			                        fmt.Println("\nProposed Build Plan:")
			                        for _, step := range manifest.Build.Steps {
			                                fmt.Printf("  - %s\n", step)
			                        }
			
			                        fmt.Println("\nExecuting build...")
			                        hash, installPath, err := ingestor.Build(cmd.Context(), manifest, url)
			                        if err != nil {
			                                return err
			                        }
			
			                        err = reg.RegisterTool(registry.Tool{
			                                Name:        manifest.Name,
			                                Source:      url,
			                                Version:     manifest.Version,
			                                LastCommit:  commit,
			                                BinaryHash:  hash,
			                                InstallPath: installPath,
			                                Type:        "source",
			                        })
						
			if err != nil {
				return err
			}

			                        fmt.Printf("\nSuccessfully ingested %s!\n", manifest.Build.Bin)

			
			return nil
		},
	}

	                                                                                                                        versionCmd = &cobra.Command{
	                                                        
	                                                                                                                                Use:   "version",
	                                                        
	                                                                                                                                Short: "Print detailed version information",
	                                                        
	                                                                                                                                Run: func(cmd *cobra.Command, args []string) {
	                                                        
	                                                                                                                                        fmt.Println("🏝️ Anyisland - Detailed Version Information")
	                                                        
	                                                                                                                                        fmt.Println("─────────────────────────────────────────────")
	                                                        
	                                                                                                                                        fmt.Printf("Version:    %s\n", cli.Version)
	                                                        
	                                                                                                                                        fmt.Printf("Commit:     %s\n", cli.GetEffectiveCommit())
	                                                        
	                                                                                                                                        fmt.Printf("Built:      %s\n", cli.GetEffectiveBuildTime())
	                                                        
	                                                                                                                                        fmt.Printf("Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	                                                        
	                                                                                                                                        fmt.Printf("Compiler:   %s\n", runtime.Version())
	                                                        
	                                                                                                                                },
	                                                        
	                                                                                                                        }
	                        

	                                                                                                                        

	                        

	                                                                        

	                        
	        	        selfInstallCmd = &cobra.Command{
	                Use:   "self-install",
	                Short: "Install Anyisland to the current system",
	                RunE: func(cmd *cobra.Command, args []string) error {
	                        sys, err := pal.New()
	                        if err != nil {
	                                return err
	                        }
	                        lm := cli.NewLifecycleManager(sys)
	                        return lm.SelfInstall()
	                },
	        }
	
	                        uninstallCmd = &cobra.Command{
	                                Use:   "uninstall [tool]",
	                                Short: "Remove a tool or Anyisland itself",
	                                Args:  cobra.MaximumNArgs(1),
	                                RunE: func(cmd *cobra.Command, args []string) error {
	                                        sys, err := pal.New()
	                                        if err != nil {
	                                                return err
	                                        }
	        
	                                        if len(args) > 0 {
	                                                toolName := args[0]
	                                                                                                                                                reg, err := registry.Open(sys.GetIslandDir())
	                                                                                                                                                if err != nil {
	                                                                                                        return err
	                                                                                                }
							defer reg.Close()
	        
	                                                tool, err := reg.GetTool(toolName)
	                                                if err != nil {
	                                                        return err
	                                                }
	                                                if tool == nil {
	                                                        return fmt.Errorf("tool %s not found", toolName)
	                                                }
	        
	                                                fmt.Printf("Uninstalling %s...\n", toolName)
	                                                
	                                                // 1. Remove binary
	                                                if tool.InstallPath != "" {
	                                                        if err := os.Remove(tool.InstallPath); err != nil {
	                                                                fmt.Printf("Warning: failed to remove binary: %v\n", err)
	                                                        }
	                                                }
	        
	                                                // 2. Remove app directory (for node/python/flutter)
	                                                appDir := filepath.Join(sys.GetIslandBinDir(), toolName+"-app")
	                                                if _, err := os.Stat(appDir); err == nil {
	                                                        if err := os.RemoveAll(appDir); err != nil {
	                                                                fmt.Printf("Warning: failed to remove app directory: %v\n", err)
	                                                        }
	                                                }
	        
	                                                // 3. Remove from registry
	                                                if err := reg.RemoveTool(toolName); err != nil {
	                                                        return err
	                                                }
	        
	                                                fmt.Printf("Successfully uninstalled %s\n", toolName)
	                                                return nil
	                                        }
	        
	                                        // Self-uninstall logic (no args)
	                                        clean, _ := cmd.Flags().GetBool("clean")
	                                        
	                                        fmt.Print("Are you sure you want to uninstall Anyisland? (y/N): ")
	                                        var response string
	                                        fmt.Scanln(&response)
	                                        if strings.ToLower(response) != "y" {
	                                                return nil
	                                        }
	        
	                                        // 1. Remove from shell
	                                        if err := pal.RemovePathFromConfig(); err != nil {
	                                                fmt.Printf("Warning: failed to remove from shell config: %v\n", err)
	                                        }
	        
	                                        // 2. Log event
	                                        lm := cli.NewLifecycleManager(sys)
	                                        lm.LogEvent(cli.LifecycleEvent{
	                                                Type:   "uninstall",
	                                                Action: "uninstall",
	                                                Status: "success",
	                                        })
	        
	                                        // 3. Optional clean
	                                        if clean {
	                                                os.RemoveAll(sys.GetIslandDir())
	                                        }
	        
	                                                                        // 4. Remove binaries from .local/bin
	                                                                        binDir := sys.GetBinDir()
	                                                                        os.Remove(filepath.Join(binDir, "anyisland"))
	                                        
	                                                                        fmt.Println("Anyisland has been uninstalled.")
	                                                                        if !clean {
	                                                                                fmt.Printf("Note: configuration and logs in %s were preserved.\n", sys.GetIslandDir())
	                                                                        }
	                                                                        return nil
	                                                                },
	                                                        }
	                                        
	                                                        daemonCmd = &cobra.Command{
	                                                                Use:   "daemon",
	                                                                Short: "Manage the Anyisland background daemon",
	                                                        }
	                                        
	                                                        daemonStartCmd = &cobra.Command{
	                                                                Use:   "daemon start",
	                                                                Short: "Start the Anyisland daemon",
	                                                                RunE: func(cmd *cobra.Command, args []string) error {
	                                                                        sys, err := pal.New()
	                                                                        if err != nil {
	                                                                                return err
	                                                                        }
	                                        
	                                                                        fmt.Println("Starting Anyisland daemon...")
	                                                                        ctx, cancel := context.WithCancel(cmd.Context())
	                                                                        defer cancel()
	                                        
	                                                                        // Setup Signal Handling
	                                                                        sigChan := make(chan os.Signal, 1)
	                                                                        signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	                                                                        go func() {
	                                                                                for sig := range sigChan {
	                                                                                                                                                                                        switch sig {
	                                                                                                                                                                                        case syscall.SIGHUP:
	                                                                                                                                                                                                fmt.Println("Received SIGHUP, performing hot-swap...")
	                                                                                                                                                                                                lm := cli.NewLifecycleManager(sys)
	                                                                                                                                                                                                if err := lm.HotSwap(""); err != nil {
	                                                                                                                                                                                                        fmt.Printf("Hot-swap failed: %v\n", err)
	                                                                                                                                                                                                }
	                                                                                                                                                                                        case syscall.SIGINT, syscall.SIGTERM:
	                                                                                                                                                                                                fmt.Println("Shutting down daemon...")
	                                                                                                                                                                                                cancel()
	                                                                                                                                                                                        }
	                                                                                }
	                                                                        }()
	                                        
	                                                                                                                                                                        reg, err := registry.Open(sys.GetIslandDir())
	                                                                                                                                                                        if err != nil {
	                                                                                                                                return err
	                                                                                                                        }
							defer reg.Close()
	                                        
	                                                                                                        broker := cli.NewUpdateBroker(sys, reg)
	                                        
	                                                                                                        go func() {
	                                        
	                                                                                                                if err := broker.Start(ctx); err != nil {
	                                        
	                                                                                                                        fmt.Printf("Broker error: %v\n", err)
	                                        
	                                                                                                                }
	                                        
	                                                                                                        }()
	                                        
	                                                                        
	                                        
	                                                                                                        // Start Discovery Server (UDP)
	                                        
	                                                                                                        srv, err := discovery.NewServer(1995)
	                                        
	                                                                                                        if err != nil {
	                                        
	                                                                                                                return fmt.Errorf("failed to start discovery server: %w", err)
	                                        
	                                                                                                        }
	                                        
	                                                                                                        defer srv.Close()
	                                        
	                                                                        
	                                        
	                                                                                                        fmt.Println("Anyisland Daemon listening on UDP :1995 and UDS :anyisland.sock")
	                                        
	                                                                        
	                                        
	                                                                                                        go func() {
	                                        
	                                                                                                                err = srv.Listen(func(p discovery.Packet) {
	                                        
	                                                                                                                        fmt.Printf("Received packet: %+v\n", p)
	                                        
	                                                                                                                        if p.Op == "REGISTER" {
	                                        
	                                                                                                                                err := reg.RegisterTool(registry.Tool{
	                                        
	                                                                                                                                        Name:    p.Name,
	                                        
	                                                                                                                                        Source:  p.Source,
	                                        
	                                                                                                                                        Version: p.Version,
	                                        
	                                                                                                                                        Type:    p.Type,
	                                        
	                                                                                                                                })
	                                        
	                                                                                                                                if err != nil {
	                                        
	                                                                                                                                        fmt.Printf("failed to register tool: %v\n", err)
	                                        
	                                                                                                                                } else {
	                                        
	                                                                                                                                        fmt.Printf("Registered tool: %s\n", p.Name)
	                                        
	                                                                                                                                }
	                                        
	                                                                                                                        }
	                                        
	                                                                                                                })
	                                        
	                                                                                                                if err != nil {
	                                        
	                                                                                                                        fmt.Printf("Discovery server error: %v\n", err)
	                                        
	                                                                                                                }
	                                        
	                                                                                                        }()
	                                        
	                                                                                                        
	                                        
	                                                                                                                                        <-ctx.Done()
	                                        
	                                                                                                        
	                                        
	                                                                                                                                        return nil
	                                        
	                                                                                                        
	                                        
	                                                                                                                                },
	                                        
	                                                                                                        
	                                        
	                                                                                                                        }
	                                        
	                                                                                                        
	                                        
	                                                                                                        
	                                        
	                                                                                                        
	                                        
	                                                                                                                        rollbackCmd = &cobra.Command{
	                                        
	                                                                                                        
	                                        
	                                                                                                                                Use:   "rollback",
	                        Short: "Rollback to the previous version of Anyisland",
	                        RunE: func(cmd *cobra.Command, args []string) error {
	                                sys, err := pal.New()
	                                if err != nil {
	                                        return err
	                                }
	                                lm := cli.NewLifecycleManager(sys)
	                                if err := lm.Rollback(); err != nil {
	                                        return err
	                                }
	                                fmt.Println("Rollback successful.")
	                                return nil
	                        },
	                }
	        
	                updateCmd = &cobra.Command{
	                        Use:   "update [tool]",
	                        Short: "Update tools or Anyisland itself",
	                        RunE: func(cmd *cobra.Command, args []string) error {
	                                sys, err := pal.New()
	                                if err != nil {
	                                        return err
	                                }
	        
	                                                                                                                        if len(args) == 0 || args[0] == "anyisland" {
	        
	                                                                                                                                lm := cli.NewLifecycleManager(sys)
	        
	                                                                                                                                ag := getSynthesizer()
	        
	                                                                                                                                
	        
	                                                                                                                                latest, available, err := lm.CheckAnyislandUpdate(cmd.Context(), ag)
	        
	                                                                                                                                if err != nil {
	        
	                                                                                                                                        return fmt.Errorf("failed to check for updates: %w", err)
	        
	                                                                                                                                }
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                        if !available {
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                if os.Getenv("ANYISLAND_JUST_UPDATED") == "1" {
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                        // If we just updated, don't say 'already up to date'
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                        if len(args) == 0 {
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                                goto updateTools
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                        }
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                        return nil
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                }
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        fmt.Printf("Anyisland is already at the latest version (%s).\n", cli.ShortCommit(latest))
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        // If no args, we just continue to update other tools
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        if len(args) == 0 {
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                goto updateTools
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        }
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        return nil
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                }
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        fmt.Printf("🔄 Updating Anyisland to %s...\n", cli.ShortCommit(latest))
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        // Perform the real update via Ingestor
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        ingestor := cli.NewIngestor(ag, sys)
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        manifest, commit, err := ingestor.Ingest(cmd.Context(), "https://github.com/nathfavour/anyisland")
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        if err != nil {
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                return err
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        }
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                                                                                
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        fmt.Printf("Downloading and building latest version (%s)...\n", cli.ShortCommit(commit))
	                                                                                                                                                                                                                                                                                                                                                                	                                                                                                
	        
	                                                                                                                                                                                                                                                                                hash, installPath, err := ingestor.Build(cmd.Context(), manifest, "https://github.com/nathfavour/anyisland")
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                if err != nil {
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                        return err
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                }
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                fmt.Println("✅ Anyisland has been updated successfully!")
	        
	                                                                                                
	        
	                                                                                                                                                                                                        
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                // Register Anyisland in the registry so it's tracked
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                reg, err := registry.Open(sys.GetIslandDir())
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                if err == nil {
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                        reg.RegisterTool(registry.Tool{
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                Name:        manifest.Name,
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                Source:      "https://github.com/nathfavour/anyisland",
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                Version:     manifest.Version,
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                LastCommit:  commit,
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                BinaryHash:  hash,
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                InstallPath: installPath,
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                Type:        "source",
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                        })
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                        reg.Close()
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                }
	        
	                                                                                                
	        
	                                                                                                                                                                                                        
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        lm.HotSwap(installPath, "ANYISLAND_JUST_UPDATED=1")	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                                return nil
	        
	                                                                                                
	        
	                                                                                                                                                                                                                                                                        }
	        
	                                                                                                
	        
	                                                                                                                                                                                                        
	        
	                                                                                                
	        
	                                                                                                                                
	        
	                                                                        
	        
	                                                                                                        updateTools:
	        
	                                                                                                        reg, err := registry.Open(sys.GetIslandDir())
	        
	                                                        
	                        if err != nil {
	                                return err
	                        }
	                        defer reg.Close()
	
	                        tools, err := reg.ListTools()
	                        if err != nil {
	                                return err
	                        }
	
	                                                          var toUpdate []registry.Tool
	                                                          if len(args) > 0 {
	                                                                  target := args[0]
	                                                                  if target == "anyisland" {
	                                                                          // Already handled above
	                                                                          return nil
	                                                                  }
	                                                                  for _, t := range tools {
	                                                                        if t.Name == target {
	                                                                        toUpdate = append(toUpdate, t)
	                                                                        break
	                                                                        }
	                                                                  }
	                                                                  if len(toUpdate) == 0 {
	                                                                        return fmt.Errorf("tool %s not found in registry", target)
	                                                                  }
	                                                                                            } else {
	                                                                                                                                                                                                                // Update all tools if no specific tool is named
	                                                                                                                                                                                                                for _, t := range tools {
	                                                                                                                                                                                                                        if !strings.HasPrefix(t.Name, "anyisland") {
	                                                                                                                                                                                                                                toUpdate = append(toUpdate, t)
	                                                                                                                                                                                                                        }
	                                                                                                                                                                                                                }	                                                                                            }
	                                                          
	                        	                        ag := getSynthesizer()
	                        ingestor := cli.NewIngestor(ag, sys)
	
	                        for _, t := range toUpdate {
	                                source := t.Source
	                                                                        if sourceFlag != "" && len(toUpdate) == 1 {
	                                                                              source = sourceFlag
	                                                                        }
	                                                                                                                fmt.Printf("Checking %s for updates...\n", t.Name)
	                                                                                                                
	                                                                                                                                                                                                                                        latestCommit, _ := ingestor.DiscoverLatestCommit(cmd.Context(), source)
	                                                                                                                                                                                                                                        if latestCommit != "" && t.LastCommit == latestCommit {
	                                                                                                                                                                                                                                                if ingestor.VerifyToolIntegrity(t.InstallPath, t.BinaryHash) {
	                                                                                                                                                                                                                                                        fmt.Printf("%s is already up-to-date (%s)\n", t.Name, cli.ShortCommit(latestCommit))
	                                                                                                                                                                                                                                                        continue
	                                                                                                                                                                                                                                                }
	                                                                                                                                                                                                                                                fmt.Printf("%s seems broken. Reinstalling...\n", t.Name)
	                                                                                                                                                                                                                                        }
	                                                                                                                                                                                                
	                                                                                                                                                                                                                                        manifest, commit, err := ingestor.Ingest(cmd.Context(), source)
	                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                        if err != nil {
	                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                      fmt.Printf("Error ingesting %s: %v\n", t.Name, err)
	                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                      continue
	                                                                                                                                                                                                                                        
	                                                                                                                                                                                                                                                                                                                }
	                                                                                                                                                                                                
	                                                                                                                                                                                                                                        if t.LastCommit != commit {
	                                                                                                                                                                                                                                                fmt.Printf("Updating %s from %s to %s...\n", t.Name, cli.ShortCommit(t.LastCommit), cli.ShortCommit(commit))
	                                                                                                                                                                                                                                        }	                                                                                                                                                
	                                                                                                                                                
	                                                                                                                                                
	                                                                                                                                                                                                                        hash, installPath, err := ingestor.Build(cmd.Context(), manifest, source)	                                
	                                                                                                        if err != nil {
	                                
	                                                                                                              fmt.Printf("Error building %s: %v\n", t.Name, err)
	                                
	                                                                                                              continue
	                                
	                                                                                                        }
	                                
	                                                                                                        reg.RegisterTool(registry.Tool{
	                                                                                                                Name:        manifest.Name,
	                                                                                                                Source:      source,
	                                                                                                                Version:     manifest.Version,
	                                                                                                                LastCommit:  commit,
	                                                                                                                BinaryHash:  hash,
	                                                                                                                InstallPath: installPath,
	                                                                                                                Type:        "source",
	                                                                                                        })
	                                	
	                                fmt.Printf("%s updated successfully!\n", t.Name)
	                        }
	
	                        return nil
	                },
	        }
	)
	func getSynthesizer() agent.Synthesizer {
	// Check if Vibeaura socket exists
	home, _ := os.UserHomeDir()
	socketPath := filepath.Join(home, ".vibeauracle", "vibeaura.sock")
	if _, err := os.Stat(socketPath); err == nil {
		return agent.NewVibeauraSynthesizer()
	}

	        return &agent.HeuristicSynthesizer{}
	}
	func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
        rootCmd.PersistentFlags().StringVarP(&sourceFlag, "source", "s", "", "Override source URL or local path")

        uninstallCmd.Flags().Bool("clean", false, "Remove all data and configurations")

        rootCmd.AddCommand(setupCmd)

        rootCmd.AddCommand(initCmd)

        rootCmd.AddCommand(listCmd)

        rootCmd.AddCommand(shellCmd)

        rootCmd.AddCommand(explainCmd)

        rootCmd.AddCommand(installCmd)

        rootCmd.AddCommand(ingestCmd)

        rootCmd.AddCommand(historyCmd)

        rootCmd.AddCommand(updateCmd)

        rootCmd.AddCommand(rollbackCmd)

        rootCmd.AddCommand(versionCmd)

        rootCmd.AddCommand(selfInstallCmd)

        rootCmd.AddCommand(uninstallCmd)

        rootCmd.AddCommand(configCmd)

        rootCmd.AddCommand(daemonCmd)

        daemonCmd.AddCommand(daemonStartCmd)

        configCmd.AddCommand(configSetCmd)

        configCmd.AddCommand(configShowCmd)

        historyCmd.AddCommand(historyRecordCmd)

        historyCmd.AddCommand(historyShowCmd)

}

