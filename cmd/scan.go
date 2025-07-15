package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wellcom-rocks/updates-sucks/pkg/config"
	"github.com/wellcom-rocks/updates-sucks/pkg/output"
	"github.com/wellcom-rocks/updates-sucks/pkg/scanner"
	"github.com/wellcom-rocks/updates-sucks/pkg/version"
)

var scanCmd = &cobra.Command{
	Use:   "scan [repository-name]",
	Short: "Scan repositories for version updates",
	Long: `Scan configured repositories for new versions. Can scan all repositories
or a specific repository by name.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		if verbose {
			fmt.Printf("Error loading config file '%s': %v\n", configFile, err)
		}
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(2) // Configuration error
	}

	// Initialize scanner
	gitScanner := scanner.NewGitScanner(verbose)

	// Determine which repositories to scan
	var reposToScan []config.Repository
	if len(args) == 1 {
		// Scan specific repository
		targetRepo := args[0]
		repo := cfg.FindRepository(targetRepo)
		if repo == nil {
			fmt.Fprintf(os.Stderr, "Repository '%s' not found in configuration\n", targetRepo)
			os.Exit(2) // Configuration error
		}
		reposToScan = []config.Repository{*repo}
	} else {
		// Scan all repositories
		reposToScan = cfg.Repositories
	}

	if !quiet {
		fmt.Printf("Scanning %d repositories...\n\n", len(reposToScan))
	}

	// Scan repositories
	var results []output.ScanResult
	hasUpdates := false
	hasErrors := false

	for _, repo := range reposToScan {
		result := output.ScanResult{
			Name:           repo.Name,
			CurrentVersion: repo.CurrentVersion,
		}

		// Get latest version
		latestVersion, err := gitScanner.GetLatestVersion(&repo)
		if err != nil {
			result.Status = "ERROR"
			result.Error = err.Error()
			hasErrors = true
			if verbose {
				fmt.Printf("Error scanning %s: %v\n", repo.Name, err)
			}
		} else {
			result.LatestVersion = latestVersion

			// Compare versions
			needsUpdate, err := compareVersions(repo.CurrentVersion, latestVersion, repo.Versioning)
			if err != nil {
				result.Status = "ERROR"
				result.Error = fmt.Sprintf("Version comparison error: %v", err)
				hasErrors = true
			} else if needsUpdate {
				result.Status = "UPDATE_AVAILABLE"
				hasUpdates = true
			} else {
				result.Status = "UP_TO_DATE"
			}
		}

		results = append(results, result)
	}

	// Output results
	jsonOutput := outputFormat == "json"
	formatter := output.NewFormatter(jsonOutput, quiet, verbose)
	formatter.PrintResults(results)

	// Determine exit code
	if hasErrors {
		os.Exit(3) // Scan error
	} else if hasUpdates {
		os.Exit(1) // Updates available
	}

	return nil // Success, no updates
}

func compareVersions(current, latest string, versioning *config.Versioning) (bool, error) {
	// Remove prefix if configured
	currentCmp := current
	latestCmp := latest

	if versioning != nil && versioning.IgnorePrefix != "" {
		currentCmp = strings.TrimPrefix(current, versioning.IgnorePrefix)
		latestCmp = strings.TrimPrefix(latest, versioning.IgnorePrefix)
	}

	// Get versioning scheme
	scheme := "semver"
	if versioning != nil && versioning.Scheme != "" {
		scheme = versioning.Scheme
	}

	// Compare versions based on scheme
	switch scheme {
	case "semver":
		result, err := version.CompareSemVer(currentCmp, latestCmp)
		if err != nil {
			return false, err
		}
		return result == version.Less, nil

	case "calver":
		result, err := version.CompareCalVer(currentCmp, latestCmp)
		if err != nil {
			return false, err
		}
		return result == version.Less, nil

	case "string":
		result, err := version.CompareString(currentCmp, latestCmp)
		if err != nil {
			return false, err
		}
		return result == version.Less, nil

	default:
		return false, fmt.Errorf("unsupported versioning scheme: %s", scheme)
	}
}
