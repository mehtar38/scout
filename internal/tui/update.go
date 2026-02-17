package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"scout/internal/ai"
	"scout/internal/scanner"
	"sort"
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
		m.searchLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.searchResults = msg.results
		m.searchActive = false
		return m, nil

	case aiQueryCompleteMsg:
		m.aiLoading = false
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

// input in dashboard view
func (m Model) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Build content to calculate total lines
	files := m.scanner.GetFiles()
	if len(files) == 0 {
		return m, nil
	}

	totalLines := 40 //estimate
	visibleHeight := m.height - 12

	switch msg.String() {
	case "up", "k":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "down", "j":
		if m.scrollOffset < totalLines-visibleHeight {
			m.scrollOffset++
		}
	case "g":
		m.scrollOffset = 0
	case "G":
		m.scrollOffset = totalLines - visibleHeight
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
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
	case "o", "enter":
		if len(items) > 0 && m.cursor < len(items) {
			item := items[m.cursor]
			//cross-platform
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", "-R", item.Location)
			case "windows":
				cmd = exec.Command("explorer", "/select,", item.Location)
			default: // linux
				// Open parent directory
				cmd = exec.Command("xdg-open", filepath.Dir(item.Location))
			}
			cmd.Start()
		}
	case "c":
		// Clear AI results and go back to normal browsing
		if m.showingAIResults {
			m.filterExt = ""
			m.filterMode = FilterAll
			m.sortMode = SortByName
			m.sortReverse = false
			m.aiResults = []scanner.Metadata{}
			m.showingAIResults = false
			m.cursor = 0
			m.scrollOffset = 0
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
		// Cycle through extensions present in current results
		// Get ALL items (not just filtered) to find all extensions
		var allItems []scanner.Metadata
		if m.showingAIResults && len(m.aiResults) > 0 {
			allItems = m.aiResults
		} else {
			allItems = m.scanner.Results // Use ALL results, not filtered
		}

		extensions := []string{""} // "" means no filter
		extSet := make(map[string]bool)

		for _, item := range allItems {
			if item.Extension != "" && !item.IsDir {
				extSet[item.Extension] = true
			}
		}

		for ext := range extSet {
			extensions = append(extensions, ext)
		}

		if len(extensions) > 1 {
			sort.Strings(extensions[1:]) // Keep "" first, sort rest
		}

		// Find current and move to next
		currentIdx := 0
		for i, ext := range extensions {
			if ext == m.filterExt {
				currentIdx = i
				break
			}
		}

		m.filterExt = extensions[(currentIdx+1)%len(extensions)]
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
	case "o":
		if len(m.searchResults) > 0 && m.cursor < len(m.searchResults) {
			result := m.searchResults[m.cursor]
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", "-R", result.Location)
			case "windows":
				cmd = exec.Command("explorer", "/select,", result.Location)
			default:
				cmd = exec.Command("xdg-open", filepath.Dir(result.Location))
			}
			if cmd != nil {
				cmd.Start()
			}
		}
	}
	return m, nil
}

// input for search
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.searchInput != "" && m.scanned {
			m.searchPattern = m.searchInput
			m.searchActive = false
			m.searchLoading = true
			m.cursor = 0
			return m, executeSearch(m.scanner, m.searchPattern, m.searchRegex, m.searchCaseSensitive)
		}
		m.searchActive = false
		return m, nil
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
	case "left", "right":
		// Just ignore them, we don't track cursor position
		return m, nil
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
	}
	return m, nil
}

// input in AI view
func (m Model) handleAIKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "/":
		// Start AI input
		m.aiActive = true
		m.aiInput = ""
		m.aiLoading = true
		return m, nil
	case "v":
		// View AI results in browser
		if len(m.aiResults) > 0 {
			m.activeTab = TabBrowser
			m.showingAIResults = true
			return m, nil
		}
	}
	return m, nil
}

// text input for AI queries
func (m Model) handleAIInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.aiInput != "" {
			m.aiQuery = m.aiInput
			m.aiActive = false
			m.aiLoading = true
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

// sent when AI query finishes
type aiQueryCompleteMsg struct {
	response string
	results  []scanner.Metadata
	err      error
}

// sends query to AI and gets results
func executeAIQuery(query string, s *scanner.Scanner) tea.Cmd {
	return func() tea.Msg {
		godotenv.Load()
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return aiQueryCompleteMsg{
				response: "Error: GEMINI_API_KEY not found in environment",
				err:      fmt.Errorf("GEMINI_API_KEY not found"),
			}
		}

		aiClient, err := ai.NewClient(apiKey)
		if err != nil {
			return aiQueryCompleteMsg{
				response: fmt.Sprintf("Error connecting to AI: %v", err),
				err:      err,
			}
		}

		parsedCmd, err := aiClient.ParseQuery(query)
		if err != nil {
			return aiQueryCompleteMsg{
				response: fmt.Sprintf("Error parsing query: %v", err),
				err:      err,
			}
		}

		if parsedCmd.Command == "none" || parsedCmd.Command == "" {
			return aiQueryCompleteMsg{
				response: "Sorry, Scout isn't able to handle this query yet. Try queries like:\n• 'show me the 10 largest files'\n• 'find all PDF files'\n• 'list recently modified files'",
				results:  []scanner.Metadata{},
				err:      nil,
			}
		}

		results := executeAICommandForTUI(parsedCmd, s)

		if len(results) == 0 {
			return aiQueryCompleteMsg{
				response: parsedCmd.Response + "\n\nNo results found matching your query.",
				results:  results,
				err:      nil,
			}
		}

		return aiQueryCompleteMsg{
			response: parsedCmd.Response,
			results:  results,
			err:      nil,
		}
	}
}

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
		if cmd.Days > 0 {
			results = s.GetFilesNewerThan(cmd.Days)
			// if cmd.Count > 0 && len(results) > cmd.Count {
			// 	results = results[:cmd.Count]
			// }
		} else {
			if cmd.Flags["dirs"] {
				results = s.GetNRecentlyModDirs(cmd.Count)
			} else {
				results = s.GetNRecentlyModFiles(cmd.Count)
			}
		}
	case "oldest":
		if cmd.Days > 0 {
			results = s.GetFilesOlderThan(cmd.Days)

			// if cmd.Count > 0 && len(results) > cmd.Count {
			// 	results = results[:cmd.Count]
			// }
		} else {
			if cmd.Flags["dirs"] {
				results = s.GetNLeastModDirs(cmd.Count)
			} else {
				results = s.GetNLeastModFiles(cmd.Count)
			}
		}
	case "search":
		opts := scanner.SearchOptions{
			Pattern:       cmd.Pattern,
			UseRegex:      cmd.Flags["regex"],
			CaseSensitive: cmd.Flags["case_sensitive"],
		}

		if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
			opts.Extensions = strings.Split(ext, ",")
		}

		searchResults := s.Search(opts)

		for _, sr := range searchResults {
			for _, item := range s.Results {
				if item.Location == sr.Location {
					results = append(results, item)
					break
				}
			}
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
