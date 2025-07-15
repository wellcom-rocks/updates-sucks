package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type ScanResult struct {
	Name           string `json:"name"`
	Status         string `json:"status"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	Error          string `json:"error,omitempty"`
}

type JSONOutput struct {
	Summary      Summary      `json:"summary"`
	Repositories []ScanResult `json:"repositories"`
}

type Summary struct {
	Total             int `json:"total"`
	UpToDate          int `json:"upToDate"`
	UpdatesAvailable  int `json:"updatesAvailable"`
	Errors            int `json:"errors"`
}

type Formatter struct {
	jsonOutput bool
	quiet      bool
	verbose    bool
}

func NewFormatter(jsonOutput, quiet, verbose bool) *Formatter {
	return &Formatter{
		jsonOutput: jsonOutput,
		quiet:      quiet,
		verbose:    verbose,
	}
}

func (f *Formatter) PrintResults(results []ScanResult) {
	if f.jsonOutput {
		f.printJSON(results)
	} else {
		f.printHuman(results)
	}
}

func (f *Formatter) printJSON(results []ScanResult) {
	summary := f.calculateSummary(results)
	
	output := JSONOutput{
		Summary:      summary,
		Repositories: results,
	}
	
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}

func (f *Formatter) printHuman(results []ScanResult) {
	summary := f.calculateSummary(results)
	
	// Print individual results
	for _, result := range results {
		switch result.Status {
		case "UP_TO_DATE":
			if !f.quiet {
				fmt.Printf("- %s: UP-TO-DATE (Current: %s)\n", result.Name, result.CurrentVersion)
			}
		case "UPDATE_AVAILABLE":
			fmt.Printf("- %s: NEW VERSION FOUND! (Current: %s -> Latest: %s)\n", 
				result.Name, result.CurrentVersion, result.LatestVersion)
		case "ERROR":
			fmt.Printf("- %s: ERROR! (%s)\n", result.Name, result.Error)
		}
	}
	
	// Print summary
	if !f.quiet {
		fmt.Printf("\nScan finished. Updates available for %d repository(ies).", summary.UpdatesAvailable)
		if summary.Errors > 0 {
			fmt.Printf(" %d error(s) occurred.", summary.Errors)
		}
		fmt.Println()
	}
}

func (f *Formatter) calculateSummary(results []ScanResult) Summary {
	summary := Summary{
		Total: len(results),
	}
	
	for _, result := range results {
		switch result.Status {
		case "UP_TO_DATE":
			summary.UpToDate++
		case "UPDATE_AVAILABLE":
			summary.UpdatesAvailable++
		case "ERROR":
			summary.Errors++
		}
	}
	
	return summary
}