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

			plan, err := ingestor.Ingest(cmd.Context(), url)
			if err != nil {
				return err
			}

			fmt.Println("\nProposed Build Plan:")
			for _, step := range plan.Steps {
				fmt.Printf("  - %s\n", step)
			}
			fmt.Printf("\nBinary target: %s\n", plan.Bin)
			
			fmt.Println("\nExecuting build...")
			if err := ingestor.Build(cmd.Context(), plan, url); err != nil {
				return err
			}

			reg, err := registry.Open(sys.GetIslandDir())
			if err != nil {
				return err
			}
			defer reg.Close()

			err = reg.RegisterTool(registry.Tool{
				Name:    plan.Bin,
				Source:  url,
				Version: "latest",
				Type:    "source",
			})
			if err != nil {
				return err
			}

			fmt.Printf("\nSuccessfully installed %s!\n", plan.Bin)
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

			plan, err := ingestor.Ingest(cmd.Context(), url)
			if err != nil {
				return err
			}

			fmt.Println("\nProposed Build Plan:")
			for _, step := range plan.Steps {
				fmt.Printf("  - %s\n", step)
			}
			fmt.Printf("\nBinary target: %s\n", plan.Bin)
			
			fmt.Println("\nExecuting build...")
			if err := ingestor.Build(cmd.Context(), plan, url); err != nil {
				return err
			}

			reg, err := registry.Open(sys.GetIslandDir())
			if err != nil {
				return err
			}
			defer reg.Close()

			err = reg.RegisterTool(registry.Tool{
				Name:    plan.Bin,
				Source:  url,
				Version: "latest",
				Type:    "source",
			})
			if err != nil {
				return err
			}

			fmt.Printf("\nSuccessfully ingested %s!\n", plan.Bin)
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
				// Default to updating anyisland if no args
				for _, t := range tools {
					if t.Name == "anyisland" {
						toUpdate = append(toUpdate, t)
						break
					}
				}
				if len(toUpdate) == 0 {
					return fmt.Errorf("anyisland not found in registry; please run 'anyisland install nathfavour/anyisland' first")
				}
			}

			ag := getSynthesizer()
			ingestor := cli.NewIngestor(ag, sys)

			for _, t := range toUpdate {
				source := t.Source
				if sourceFlag != "" && len(toUpdate) == 1 {
					source = sourceFlag
				}
				fmt.Printf("Updating %s from %s...\n", t.Name, source)
				
				plan, err := ingestor.Ingest(cmd.Context(), source)
				if err != nil {
					fmt.Printf("Error ingesting %s: %v\n", t.Name, err)
					continue
				}

				if err := ingestor.Build(cmd.Context(), plan, source); err != nil {
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
	sys, err := pal.New()
	if err != nil {
		return &agent.MockSynthesizer{}
	}

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

	rootCmd.AddCommand(setupCmd)

	rootCmd.AddCommand(initCmd)

	rootCmd.AddCommand(listCmd)

	rootCmd.AddCommand(installCmd)

	rootCmd.AddCommand(ingestCmd)

	rootCmd.AddCommand(historyCmd)

	rootCmd.AddCommand(updateCmd)

	historyCmd.AddCommand(historyRecordCmd)

	historyCmd.AddCommand(historyShowCmd)

}
