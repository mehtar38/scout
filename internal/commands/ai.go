package commands

import (
	"fmt"
	"os"
	"scout/internal/ai"
	"scout/internal/scanner"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai <query>",
	Short: "Natural language file queries powered by AI",
	Long:  `Ask questions about your files in natural language. Prefix with ~ai in any command.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   runAI,
}

func init() {
	rootCmd.AddCommand(aiCmd)
}

func runAI(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error getting API Key")
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

	fmt.Println("Success!", parsedCmd.Response)
	fmt.Println()

	executeAICommand(parsedCmd)
}

func executeAICommand(cmd *ai.CommandParser) {
	path := "."

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

	// Execute based on command type
	var results []scanner.Metadata

	switch cmd.Command {
	case "largest":
		results = s.GetNLargestFilesBySize(cmd.Count)
	case "smallest":
		results = s.GetNSmallestFilesBySize(cmd.Count)
	case "recent":
		results = s.GetNRecentlyModFiles(cmd.Count)
	case "oldest":
		results = s.GetNLeastModFiles(cmd.Count)
	default:
		fmt.Println("Unknown command:", cmd.Command)
		return
	}

	results = applyFilters(results, cmd.Filters)

	displayAIResults(results)
}

func applyFilters(files []scanner.Metadata, filters map[string]string) []scanner.Metadata {
	if len(filters) == 0 {
		return files
	}

	filtered := []scanner.Metadata{}

	for _, file := range files {
		include := true

		// Filter by extension
		if ext, ok := filters["extension"]; ok && ext != "" {
			// Handle comma-separated extensions like ".mp4,.avi,.mkv"
			extensions := strings.Split(ext, ",")
			matched := false
			for _, e := range extensions {
				e = strings.TrimSpace(e)
				if file.Extension == e {
					matched = true
					break
				}
			}
			if !matched {
				include = false
			}
		}

		// Filter by older_than_days
		if olderStr, ok := filters["older_than_days"]; ok && olderStr != "" {
			days, err := strconv.Atoi(olderStr)
			if err == nil {
				cutoff := time.Now().AddDate(0, 0, -days)
				if file.ModificationTime.After(cutoff) {
					include = false
				}
			}
		}

		// Filter by newer_than_days
		if newerStr, ok := filters["newer_than_days"]; ok && newerStr != "" {
			days, err := strconv.Atoi(newerStr)
			if err == nil {
				cutoff := time.Now().AddDate(0, 0, -days)
				if file.ModificationTime.Before(cutoff) {
					include = false
				}
			}
		}

		if include {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

func displayAIResults(results []scanner.Metadata) {
	if len(results) == 0 {
		fmt.Println("No files found matching your criteria.")
		return
	}

	fmt.Printf("Found %d files:\n\n", len(results))
	for _, file := range results {
		fmt.Printf("📄 %s\n", file.Name)
		fmt.Printf("   Location: %s\n", file.Location)
		fmt.Printf("   Size: %d bytes\n", file.Size)
		fmt.Printf("   Modified: %s\n\n", file.ModificationTime.Format("2006-01-02 15:04"))
	}
}
