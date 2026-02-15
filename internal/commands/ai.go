package commands

import (
	"fmt"
	"os"
	"scout/internal/ai"
	"scout/internal/scanner"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai <query>",
	Short: "Natural language file queries powered by AI",
	Args:  cobra.MinimumNArgs(1),
	Run:   runAI,
}

func init() {
	rootCmd.AddCommand(aiCmd)
	godotenv.Load()
}

func runAI(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error fetching API key")
		return
	}

	aiClient, err := ai.NewClient(apiKey)
	if err != nil {
		fmt.Println("Error creating AI client:", err)
		return
	}

	parsedCmd, err := aiClient.ParseQuery(query)
	if err != nil {
		fmt.Println("Error parsing query:", err)
		return
	}

	fmt.Println(parsedCmd.Response)
	fmt.Println()

	executeAICommand(parsedCmd)
}

func executeAICommand(cmd *ai.CommandParser) {
	path := cmd.Path
	if path == "" {
		path = "."
	}

	// Scan directory
	s, err := scanner.New(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := s.Scan(); err != nil {
		fmt.Println("Error scanning:", err)
		return
	}

	s.ComputeDirSize()

	// Route to appropriate command using EXISTING logic
	switch cmd.Command {
	case "list":
		executeListCommand(s, cmd)
	case "largest":
		executeLargestCommand(s, cmd)
	case "smallest":
		executeSmallestCommand(s, cmd)
	case "recent":
		executeRecentCommand(s, cmd)
	case "oldest":
		executeOldestCommand(s, cmd)
	case "search":
		executeSearchCommand(s, cmd)
	case "stats":
		executeStatsCommand(s, cmd)
	default:
		fmt.Println("Unknown command:", cmd.Command)
	}
}

func executeLargestCommand(s *scanner.Scanner, cmd *ai.CommandParser) {
	var results []scanner.Metadata

	if cmd.Flags["files"] {
		results = s.GetNLargestFilesBySize(cmd.Count)
	} else if cmd.Flags["dirs"] {
		results = s.GetNLargestDirectoriesBySize(cmd.Count)
	} else {
		results = s.GetNLargestFilesBySize(cmd.Count) // both by default
	}

	if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
		results = filterByExtension(results, ext)
	}

	displayResults(results)
}

func executeSmallestCommand(s *scanner.Scanner, cmd *ai.CommandParser) {
	var results []scanner.Metadata

	if cmd.Flags["files"] {
		results = s.GetNSmallestFilesBySize(cmd.Count)
	} else if cmd.Flags["dirs"] {
		results = s.GetNSmallestDirectoriesBySize(cmd.Count)
	} else {
		results = s.GetNSmallestFilesBySize(cmd.Count)
	}

	if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
		results = filterByExtension(results, ext)
	}

	displayResults(results)
}

func executeRecentCommand(s *scanner.Scanner, cmd *ai.CommandParser) {
	var results []scanner.Metadata

	if cmd.Flags["dirs"] {
		results = s.GetNRecentlyModDirs(cmd.Count)
	} else {
		results = s.GetNRecentlyModFiles(cmd.Count)
	}

	if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
		results = filterByExtension(results, ext)
	}

	displayResults(results)
}

func executeOldestCommand(s *scanner.Scanner, cmd *ai.CommandParser) {
	var results []scanner.Metadata

	if cmd.Flags["dirs"] {
		results = s.GetNLeastModDirs(cmd.Count)
	} else {
		results = s.GetNLeastModFiles(cmd.Count)
	}

	if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
		results = filterByExtension(results, ext)
	}

	displayResults(results)
}

func executeListCommand(s *scanner.Scanner, cmd *ai.CommandParser) {
	var results []scanner.Metadata

	if cmd.Flags["files"] {
		results = s.GetFiles()
	} else if cmd.Flags["dirs"] {
		results = s.GetDirectories()
	} else {
		results = s.Results
	}

	if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
		results = filterByExtension(results, ext)
	}

	displayResults(results)
}

func executeSearchCommand(s *scanner.Scanner, cmd *ai.CommandParser) {
	opts := scanner.SearchOptions{
		Pattern:       cmd.Pattern,
		UseRegex:      cmd.Flags["regex"],
		CaseSensitive: cmd.Flags["case_sensitive"],
	}

	if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
		opts.Extensions = strings.Split(ext, ",")
	}

	results := s.Search(opts)

	if len(results) == 0 {
		fmt.Println("No matches found.")
		return
	}

	fmt.Printf("Found matches in %d files:\n\n", len(results))

	for _, result := range results {
		if cmd.Flags["verbose"] {
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

func executeStatsCommand(s *scanner.Scanner, _ *ai.CommandParser) {
	// files := s.GetFiles()
	// dirs := s.GetDirectories()
	// var results []scanner.Metadata
	// path := cmd.Path
	s.ComputeDirSize()
	s.GetStats()
}

// Helper functions
func filterByExtension(files []scanner.Metadata, extensions string) []scanner.Metadata {
	exts := strings.Split(extensions, ",")
	for i := range exts {
		exts[i] = strings.TrimSpace(exts[i])
	}

	filtered := []scanner.Metadata{}
	for _, file := range files {
		for _, ext := range exts {
			if file.Extension == ext {
				filtered = append(filtered, file)
				break
			}
		}
	}
	return filtered
}

func displayResults(results []scanner.Metadata) {
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("Found %d items:\n\n", len(results))
	for _, item := range results {
		fmt.Printf("📄 %s\n", item.Name)
		fmt.Printf("   Location: %s\n", item.Location)
		if item.IsDir {
			fmt.Printf("   Size: %d bytes (total)\n", item.DirSize)
		} else {
			fmt.Printf("   Size: %d bytes\n", item.Size)
		}
		fmt.Printf("   Modified: %s\n\n", item.ModificationTime.Format("2006-01-02 15:04"))
	}
}
