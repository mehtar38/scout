package commands

import (
	"fmt"

	"github.com/mehtar38/scout/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui [path]",
	Short: "Launch interactive TUI dashboard",
	Long:  `Launch an interactive terminal UI for browsing and analyzing files`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	m := tui.NewModel(path)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
	}
}
