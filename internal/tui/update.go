package tui

import (
	"fmt"
	"os"
	"scout/internal/ai"
	"scout/internal/scanner"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case scanCompleteMsg:
		m.scanning = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.scanner = msg.scanner
		m.scanned = true
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil

	case searchCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.searchResults = msg.results
		m.searchActive = false
		return m, nil

	case aiQueryCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.aiResponse = msg.response
		m.aiResults = msg.results
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	//Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if !m.searchActive && !m.aiActive {
			return m, tea.Quit
		}
	case "esc":
		if m.searchActive {
			m.searchActive = false
			m.searchInput = ""
			return m, nil
		}
		if m.aiActive {
			m.aiActive = false
			m.aiInput = ""
			return m, nil
		}
	}

	// If we're in input mode, handle text input
	if m.searchActive {
		return m.handleSearchInput(msg)
	}
	if m.aiActive {
		return m.handleAIInput(msg)
	}

	// Tab navigation (when not in input mode)
	switch msg.String() {
	case "1":
		m.activeTab = TabDashboard
		m.cursor = 0
		return m, nil
	case "2":
		m.activeTab = TabBrowser
		m.cursor = 0
		return m, nil
	case "3":
		m.activeTab = TabSearch
		m.cursor = 0
		return m, nil
	case "4":
		m.activeTab = TabAI
		m.cursor = 0
		return m, nil
	case "tab":
		m.activeTab = (m.activeTab + 1) % 4
		m.cursor = 0
		return m, nil
	case "shift+tab":
		m.activeTab = (m.activeTab - 1 + 4) % 4
		m.cursor = 0
		return m, nil
	}

	if msg.String() == "r" || msg.String() == "f5" {
		m.scanning = true
		return m, scanDirectory(m.path)
	}

	// Tab-specific controls
	switch m.activeTab {
	case TabDashboard:
		return m.handleDashboardKeys(msg)
	case TabBrowser:
		return m.handleBrowserKeys(msg)
	case TabSearch:
		return m.handleSearchKeys(msg)
	case TabAI:
		return m.handleAIKeys(msg)
	}

	return m, nil
}

// handleDashboardKeys handles input in dashboard view
func (m Model) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Dashboard is mostly read-only, just navigation
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		m.cursor++
	}
	return m, nil
}

// input in browser view
func (m Model) handleBrowserKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.scanned {
		return m, nil
	}

	items := m.getBrowserItems()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollOffset {
				m.scrollOffset = m.cursor
			}
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
			visibleHeight := m.height - 10 // Account for header/footer
			if m.cursor >= m.scrollOffset+visibleHeight {
				m.scrollOffset = m.cursor - visibleHeight + 1
			}
		}
	case "g":
		m.cursor = 0
		m.scrollOffset = 0
	case "G":
		m.cursor = len(items) - 1
		visibleHeight := m.height - 10
		if len(items) > visibleHeight {
			m.scrollOffset = len(items) - visibleHeight
		}
	case "s":
		// Cycle sort mode
		m.sortMode = (m.sortMode + 1) % 4
		m.cursor = 0
		m.scrollOffset = 0
	case "S":
		// Reverse sort
		m.sortReverse = !m.sortReverse
		m.cursor = 0
		m.scrollOffset = 0
	case "f":
		// Cycle filter mode
		m.filterMode = (m.filterMode + 1) % 3
		m.cursor = 0
		m.scrollOffset = 0
	case "e":
		// Toggle extension filter (placeholder - would need input)
		if m.filterExt == "" {
			m.filterExt = ".go" // Example
		} else {
			m.filterExt = ""
		}
		m.cursor = 0
		m.scrollOffset = 0
	}

	return m, nil
}

// handleSearchKeys handles input in search view
func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "/":
		// Start search input
		m.searchActive = true
		m.searchInput = ""
		return m, nil
	case "r":
		// Toggle regex
		m.searchRegex = !m.searchRegex
		return m, nil
	case "c":
		// Toggle case sensitivity
		m.searchCaseSensitive = !m.searchCaseSensitive
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.searchResults)-1 {
			m.cursor++
			visibleHeight := m.height - 18
			if m.cursor >= m.scrollOffset+visibleHeight {
				m.scrollOffset = m.cursor - visibleHeight + 1
			}
		}
	}
	return m, nil
}

// handleSearchInput handles text input for search
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.searchInput != "" && m.scanned {
			m.searchPattern = m.searchInput
			m.searchActive = false
			m.cursor = 0
			return m, executeSearch(m.scanner, m.searchPattern, m.searchRegex, m.searchCaseSensitive)
		}
		m.searchActive = false
		return m, nil
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
	}
	return m, nil
}

// handleAIKeys handles input in AI view
func (m Model) handleAIKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "/":
		// Start AI input
		m.aiActive = true
		m.aiInput = ""
		return m, nil
	case "v":
		// View AI results in browser
		if len(m.aiResults) > 0 {
			m.activeTab = TabBrowser
			return m, nil
		}
	}
	return m, nil
}

// handleAIInput handles text input for AI queries
func (m Model) handleAIInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.aiInput != "" {
			m.aiQuery = m.aiInput
			m.aiActive = false
			return m, executeAIQuery(m.aiInput, m.scanner)
		}
		m.aiActive = false
		return m, nil
	case "backspace":
		if len(m.aiInput) > 0 {
			m.aiInput = m.aiInput[:len(m.aiInput)-1]
		}
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.aiInput += msg.String()
		}
	}
	return m, nil
}

// aiQueryCompleteMsg is sent when AI query finishes
type aiQueryCompleteMsg struct {
	response string
	results  []scanner.Metadata
	err      error
}

// executeAIQuery sends query to AI and gets results
func executeAIQuery(query string, s *scanner.Scanner) tea.Cmd {
	return func() tea.Msg {
		godotenv.Load()
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return aiQueryCompleteMsg{err: fmt.Errorf("GEMINI_API_KEY not found")}
		}

		aiClient, err := ai.NewClient(apiKey)
		if err != nil {
			return aiQueryCompleteMsg{err: err}
		}

		parsedCmd, err := aiClient.ParseQuery(query)
		if err != nil {
			return aiQueryCompleteMsg{err: err}
		}

		// ADD THIS DEBUG LINE
		fmt.Fprintf(os.Stderr, "DEBUG: Command=%s, Count=%d, Flags=%+v\n",
			parsedCmd.Command, parsedCmd.Count, parsedCmd.Flags)

		results := executeAICommandForTUI(parsedCmd, s)

		// ADD THIS TOO
		fmt.Fprintf(os.Stderr, "DEBUG: Got %d results\n", len(results))

		return aiQueryCompleteMsg{
			response: parsedCmd.Response,
			results:  results,
			err:      nil,
		}
	}
}

// executeAICommandForTUI executes AI command and returns results
func executeAICommandForTUI(cmd *ai.CommandParser, s *scanner.Scanner) []scanner.Metadata {
	var results []scanner.Metadata

	switch cmd.Command {
	case "list":
		if cmd.Flags["files"] {
			results = s.GetFiles()
		} else if cmd.Flags["dirs"] {
			results = s.GetDirectories()
		} else {
			results = s.Results
		}
	case "largest":
		if cmd.Flags["files"] {
			results = s.GetNLargestFilesBySize(cmd.Count)
		} else if cmd.Flags["dirs"] {
			results = s.GetNLargestDirectoriesBySize(cmd.Count)
		} else {
			results = s.GetNLargestFilesBySize(cmd.Count)
		}
	case "smallest":
		if cmd.Flags["files"] {
			results = s.GetNSmallestFilesBySize(cmd.Count)
		} else if cmd.Flags["dirs"] {
			results = s.GetNSmallestDirectoriesBySize(cmd.Count)
		} else {
			results = s.GetNSmallestFilesBySize(cmd.Count)
		}
	case "recent":
		if cmd.Flags["dirs"] {
			results = s.GetNRecentlyModDirs(cmd.Count)
		} else {
			results = s.GetNRecentlyModFiles(cmd.Count)
		}
	case "oldest":
		if cmd.Flags["dirs"] {
			results = s.GetNLeastModDirs(cmd.Count)
		} else {
			results = s.GetNLeastModFiles(cmd.Count)
		}
	}

	// Apply extension filter if present
	if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
		results = filterByExtension(results, ext)
	}

	return results
}

// filterByExtension filters results by extension
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

// Update handler for AI query completion
func (m Model) updateAIComplete(msg aiQueryCompleteMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}
	m.aiResponse = msg.response
	m.aiResults = msg.results
	return m, nil
}
