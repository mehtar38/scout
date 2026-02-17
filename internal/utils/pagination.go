package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const defaultPageSize = 20

func Paginate(lines []string, pageSize int) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	totalLines := len(lines)
	if totalLines <= pageSize {
		for _, line := range lines {
			fmt.Println(line)
		}
		return
	}

	reader := bufio.NewReader(os.Stdin)
	currentPage := 0
	totalPages := (totalLines + pageSize - 1) / pageSize

	for {
		start := currentPage * pageSize
		end := start + pageSize
		if end > totalLines {
			end = totalLines
		}

		for i := start; i < end; i++ {
			fmt.Println(lines[i])
		}

		fmt.Printf("\n--- Page %d of %d  ---\n",
			currentPage+1, totalPages)
		fmt.Print("[Enter: next | b: back | q: quit] > ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q":
			return
		case "b":
			if currentPage > 0 {
				currentPage--
			}
		case "":
			currentPage++
			if currentPage >= totalPages {
				fmt.Println("\nEnd of results.")
				return
			}
		default:
			fmt.Println("Invalid input. Use Enter, 'b', or 'q'")
		}

		// Clear screen for next page
		fmt.Print("\033[H\033[2J")
	}
}

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
