package main

import (
        "fmt"
        "os"
        "path/filepath"
        "strings"
	"github.com/spf13/cobra"
	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/nathfavour/anyisland/internal/registry"
	"github.com/nathfavour/anyisland/internal/cli"
	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/history"
	"github.com/nathfavour/anyisland/internal/crypto"
)

var (
	sourceFlag string

	rootCmd = &cobra.Command{
		Use:   "anyisland",
		Short: "Anyisland is an AI-powered package manager",
		Long:  `Anyisland is an AI-powered, platform-agnostic, and decentralized package manager.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip check for setup and init command
			if cmd.Name() == "setup" || cmd.Name() == "init" {
				return nil
			}
			sys, err := pal.New()
			if err != nil {
				return nil // Ignore PAL errors in PreRun to allow emergency use
			}
			return sys.EnsurePath()
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
			if err := sys.InjectPath(); err != nil {
				fmt.Printf("Warning: failed to inject PATH: %v\n", err)
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
				return fmt.Errorf("must provide a URL or use --source")
			}
			sys, err := pal.New()
			if err != nil {
				return err
			}

			ag := getSynthesizer()
			ingestor := cli.NewIngestor(ag, sys)

			                        manifest, err := ingestor.Ingest(cmd.Context(), url)

			                        if err != nil {

			                                return err

			                        }

			

			                        fmt.Println("\nProposed Build Plan:")

			                        for _, step := range manifest.Build.Steps {

			                                fmt.Printf("  - %s\n", step)

			                        }

			                        fmt.Printf("\nBinary target: %s\n", manifest.Build.Bin)

			

			                        fmt.Println("\nExecuting build...")

			                        if err := ingestor.Build(cmd.Context(), manifest, url); err != nil {

			                                return err

			                        }

			

			                        reg, err := registry.Open(sys.GetIslandDir())

			                        if err != nil {

			                                return err

			                        }

			                        defer reg.Close()

			

			                        err = reg.RegisterTool(registry.Tool{

			                                Name:    manifest.Build.Bin,

			                                Source:  url,

			                                Version: manifest.Version,

			                                Type:    "source",

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

			                        manifest, err := ingestor.Ingest(cmd.Context(), url)

			                        if err != nil {

			                                return err

			                        }

			

			                        fmt.Println("\nProposed Build Plan:")

			                        for _, step := range manifest.Build.Steps {

			                                fmt.Printf("  - %s\n", step)

			                        }

			                        fmt.Printf("\nBinary target: %s\n", manifest.Build.Bin)

			

			                        fmt.Println("\nExecuting build...")

			                        if err := ingestor.Build(cmd.Context(), manifest, url); err != nil {

			                                return err

			                        }

			

			                        reg, err := registry.Open(sys.GetIslandDir())

			                        if err != nil {

			                                return err

			                        }

			                        defer reg.Close()

			

			                        err = reg.RegisterTool(registry.Tool{

			                                Name:    manifest.Build.Bin,

			                                Source:  url,

			                                Version: manifest.Version,

			                                Type:    "source",

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
	                Short: "Print version information",
	                Run: func(cmd *cobra.Command, args []string) {
	                        fmt.Println(cli.VersionString())
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
	                Use:   "uninstall",
	                Short: "Remove Anyisland from the system",
	                RunE: func(cmd *cobra.Command, args []string) error {
	                        clean, _ := cmd.Flags().GetBool("clean")
	                        sys, err := pal.New()
	                        if err != nil {
	                                return err
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
	                        os.Remove(filepath.Join(binDir, "anyislandd"))
	
	                        fmt.Println("Anyisland has been uninstalled.")
	                        if !clean {
	                                fmt.Printf("Note: configuration and logs in %s were preserved.\n", sys.GetIslandDir())
	                        }
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
	                                        fmt.Println("Checking for Anyisland updates...")
	                                        // In a real scenario, we'd check remote version/checksum here
	                                        
	                                        // Simulation of backup for rollback demonstration
	                                        binPath := filepath.Join(sys.GetBinDir(), "anyisland")
	                                        os.Rename(binPath, binPath+".old")
	                                        // Copy current to binPath as if it were a new download
	                                        exePath, _ := os.Executable()
	                                        data, _ := os.ReadFile(exePath)
	                                        os.WriteFile(binPath, data, 0755)
	        
	                                        fmt.Println("Anyisland updated to latest version (simulated).")
	                                        return nil
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
	
	                        var toUpdate []registry.Tool
	                        target := args[0]
	                        for _, t := range tools {
	                              if t.Name == target {
	                              toUpdate = append(toUpdate, t)
	                              break
	                              }
	                        }
	                        if len(toUpdate) == 0 {
	                              return fmt.Errorf("tool %s not found in registry", target)
	                        }
	
	                        ag := getSynthesizer()
	                        ingestor := cli.NewIngestor(ag, sys)
	
	                        for _, t := range toUpdate {
	                                source := t.Source
	                                if sourceFlag != "" && len(toUpdate) == 1 {
	                                      source = sourceFlag
	                                }
	                                fmt.Printf("Updating %s from %s...\n", t.Name, source)
	
	                                                                manifest, err := ingestor.Ingest(cmd.Context(), source)
	
	                                                                if err != nil {
	
	                                                                      fmt.Printf("Error ingesting %s: %v\n", t.Name, err)
	
	                                                                      continue
	
	                                                                }
	
	
	
	                                                                if err := ingestor.Build(cmd.Context(), manifest, source); err != nil {
	
	                                                                      fmt.Printf("Error building %s: %v\n", t.Name, err)
	
	                                                                      continue
	
	                                                                }
	
	
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

	return &agent.MockSynthesizer{}
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

        rootCmd.AddCommand(installCmd)

        rootCmd.AddCommand(ingestCmd)

        rootCmd.AddCommand(historyCmd)

        rootCmd.AddCommand(updateCmd)

        rootCmd.AddCommand(rollbackCmd)

        rootCmd.AddCommand(versionCmd)

        rootCmd.AddCommand(selfInstallCmd)

        rootCmd.AddCommand(uninstallCmd)

        historyCmd.AddCommand(historyRecordCmd)

        historyCmd.AddCommand(historyShowCmd)

}

