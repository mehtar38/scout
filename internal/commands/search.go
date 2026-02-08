package commands

import (
	"fmt"
	"scout/internal/scanner"
	"strings"

	"github.com/spf13/cobra"
)

var (
	searchRegex         bool
	searchCaseSensitive bool
	searchExtensions    string
	searchVerbose       bool
)

var searchCmd = &cobra.Command{
	Use:   "search <pattern> [path]",
	Short: "Search for text patterns in files",
	Long:  `Search through files for a text pattern or regular expression.`,
	Args:  cobra.RangeArgs(1, 2),
	Run:   runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&searchRegex, "regex", false, "Use regex pattern matching")
	searchCmd.Flags().BoolVar(&searchCaseSensitive, "i", false, "Case-sensitive search")
	searchCmd.Flags().StringVar(&searchExtensions, "ext", "", "Comma-separated list of extensions (e.g., '.go,.txt')")
	searchCmd.Flags().BoolVarP(&searchVerbose, "verbose", "v", false, "Show line numbers for each match")

	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) {
	pattern := args[0]
	path := "."
	if len(args) > 1 {
		path = args[1]
	}

	s, err := scanner.New(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = s.Scan()
	if err != nil {
		fmt.Println("Error scanning:", err)
		return
	}

	// Parse extensions
	var extensions []string
	if searchExtensions != "" {
		extensions = strings.Split(searchExtensions, ",")
	}

	// Build search options
	opts := scanner.SearchOptions{
		Pattern:       pattern,
		Extensions:    extensions,
		UseRegex:      searchRegex,
		CaseSensitive: searchCaseSensitive,
	}

	results := s.Search(opts)

	if len(results) == 0 {
		fmt.Println("No matches found.")
		return
	}

	fmt.Printf("Found matches in %d files:\n\n", len(results))

	for _, result := range results {
		if searchVerbose {
			fmt.Printf("%s: %d matches\n", result.Location, result.Matches)
			for _, lineNum := range result.LineNumbers {
				fmt.Printf("  Line %d\n", lineNum)
			}
			fmt.Println()
		} else {
			fmt.Printf("%s: %d matches\n", result.Location, result.Matches)
		}
	}
}
