package main

import (
	"fmt"
	"os"
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
	rootCmd = &cobra.Command{
		Use:   "anyisland",
		Short: "Anyisland is an AI-powered package manager",
		Long:  `Anyisland is an AI-powered, platform-agnostic, and decentralized package manager.`,
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
			ag := &agent.MockSynthesizer{}
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
			ag := &agent.MockSynthesizer{}
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

	initCmd = &cobra.Command{
		Use:   "init",
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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			sys, err := pal.New()
			if err != nil {
				return err
			}

			ag := &agent.MockSynthesizer{}
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
		Use:   "update",
		Short: "Update Anyisland itself",
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

			var self *registry.Tool
			for _, t := range tools {
				if t.Name == "anyisland" {
					self = &t
					break
				}
			}

			if self == nil {
				return fmt.Errorf("anyisland not found in registry; please run 'anyisland ingest github.com/nathfavour/anyisland' first")
			}

			fmt.Printf("Updating Anyisland from %s...\n", self.Source)
			ag := &agent.MockSynthesizer{}
			ingestor := cli.NewIngestor(ag, sys)

			plan, err := ingestor.Ingest(cmd.Context(), self.Source)
			if err != nil {
				return err
			}

			if err := ingestor.Build(cmd.Context(), plan, self.Source); err != nil {
				return err
			}

			fmt.Println("\nAnyisland updated successfully!")
			return nil
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(updateCmd)
	historyCmd.AddCommand(historyRecordCmd)
	historyCmd.AddCommand(historyShowCmd)
}