package cmd

import (
	"github.com/spf13/cobra"
)

var (
	configFile   string
	verbose      bool
	quiet        bool
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "version-scanner",
	Short: "A CLI tool for monitoring software version updates",
	Long: `Version-Scanner is a CLI tool designed to automate software version monitoring
for DevOps engineers and system administrators. It scans configured repositories
for new versions and integrates with CI/CD pipelines.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "file", "repos.json", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Enable quiet output")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "human", "Output format (human, json)")
}