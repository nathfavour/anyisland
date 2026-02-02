package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/nathfavour/anyisland/internal/registry"
)

var rootCmd = &cobra.Command{
	Use:   "anyisland",
	Short: "Anyisland is an AI-powered package manager",
	Long:  `Anyisland is an AI-powered, platform-agnostic, and decentralized package manager.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(listCmd)
}

var initCmd = &cobra.Command{
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
		_, err = registry.Open(sys.GetIslandDir())
		if err != nil {
			return err
		}
		fmt.Printf("Anyisland initialized at %s\n", sys.GetIslandDir())
		return nil
	},
}

var listCmd = &cobra.Command{
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