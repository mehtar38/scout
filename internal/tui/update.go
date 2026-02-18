package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mehtar38/scout/internal/ai"
	"github.com/mehtar38/scout/internal/scanner"
	"github.com/mehtar38/scout/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// INTERCEPTOR 1: If typing a question
	if m.fileQuestionInputActive {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				query := m.fileQuestionInput
				m.fileQuestionInputActive = false
				m.fileQuestionInput = ""
				m.fileActionLoading = true
				return m, performFileAction(m.selectedFile.Location, "question", query) // Issue #4 fixed
			case "esc":
				m.fileQuestionInputActive = false
				return m, nil
			case "backspace":
				if len(m.fileQuestionInput) > 0 {
					m.fileQuestionInput = m.fileQuestionInput[:len(m.fileQuestionInput)-1]
				}
			default:
				if len(key.String()) == 1 {
					m.fileQuestionInput += key.String()
				}
			}
		}
		return m, nil
	}

	// INTERCEPTOR 2: If viewing a result (Fixes Issue #3)
	if m.showActionResult {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up":
				if m.fileActionScrollOffset > 0 {
					m.fileActionScrollOffset--
				}
			case "down":
				m.fileActionScrollOffset++
			case "esc":
				m.showActionResult = false
				m.fileActionScrollOffset = 0
			}
		}
		return m, nil
	}

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

	case fileActionCompleteMsg:
		m.fileActionLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.fileActionResult = msg.result
		m.showActionResult = true
		m.fileActionScrollOffset = 1
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	//Global keys

	if m.showActionMenu {
		return m.handleActionMenuKeys(msg)
	}
	if m.showActionResult {
		return m.handleActionResultKeys(msg)
	}

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
	case "a":
		//ai actions
		if m.showingAIResults && len(m.aiResults) > 0 {
			items := m.getBrowserItems()
			if m.cursor < len(items) {
				m.selectedFile = &items[m.cursor]
				m.showActionMenu = true
				m.actionMenuCursor = 0
			}
		}
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

func (m Model) handleActionMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.actionMenuCursor > 0 {
			m.actionMenuCursor--
		}
	case "down", "j":
		if m.actionMenuCursor < 4 {
			m.actionMenuCursor++
		}
	case "enter":
		actions := []string{"Summarize document", "Extract key points", "Ask a question", "Find information", "Cancel"}
		choice := actions[m.actionMenuCursor]

		if choice == "Cancel" {
			m.showActionMenu = false
			return m, nil
		}

		if choice == "Ask a question" {
			m.showActionMenu = false
			m.fileQuestionInputActive = true // Transitions to typing mode
			return m, nil
		}

		// For actions that don't need typing:
		m.showActionMenu = false
		m.fileActionLoading = true

		actionMap := map[string]string{
			"Summarize document": "summarize",
			"Extract key points": "keypoints",
			"Find information":   "findinfo",
		}
		return m, performFileAction(m.selectedFile.Location, actionMap[choice], "")

	case "esc":
		m.showActionMenu = false
		m.selectedFile = nil
	}
	return m, nil
}

func (m Model) handleActionResultKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "enter":
		m.showActionResult = false
		m.fileActionScrollOffset = 0
		m.selectedFile = nil
		m.fileActionResult = ""
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
			// Check if we have a selected file (question mode)
			if m.selectedFile != nil {
				// This is a question about the file
				m.aiActive = false
				m.fileActionLoading = true
				return m, performFileQuestion(m.selectedFile.Location, m.aiInput)
			}

			// Normal AI query
			query := m.aiInput
			m.aiQuery = query
			m.aiInput = ""
			m.aiActive = false
			m.aiLoading = true
			return m, executeAIQuery(query, m.scanner)
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

func performFileQuestion(filepath, question string) tea.Cmd {
	return func() tea.Msg {
		var content string
		var err error

		// Read file (PDF or text)
		if utils.IsPDF(filepath) {
			content, err = utils.ExtractPDFText(filepath)
		} else if utils.IsTextFile(filepath) {
			contentBytes, readErr := os.ReadFile(filepath)
			if readErr != nil {
				return fileActionCompleteMsg{err: fmt.Errorf("could not read file: %v", readErr)}
			}
			content = string(contentBytes)
		} else {
			return fileActionCompleteMsg{err: fmt.Errorf("unsupported file type")}
		}

		if err != nil {
			return fileActionCompleteMsg{err: err}
		}

		// Truncate if needed
		const maxChars = 100000
		if len(content) > maxChars {
			content = content[:maxChars] + "\n\n[Content truncated...]"
		}

		prompt := fmt.Sprintf(`Answer this question about the document. Provide a clear, concise answer.

Question: %s

Document:
%s`, question, content)

		// Call AI
		godotenv.Load()
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fileActionCompleteMsg{err: fmt.Errorf("GEMINI_API_KEY not found")}
		}

		aiClient, err := ai.NewClient(apiKey)
		if err != nil {
			return fileActionCompleteMsg{err: err}
		}

		result, err := aiClient.ProcessDocument(prompt)
		if err != nil {
			return fileActionCompleteMsg{err: err}
		}

		return fileActionCompleteMsg{result: result}
	}
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

		fmt.Fprintf(os.Stderr, "DEBUG: AI returned - Command=%s, Pattern=%s, Filters=%+v\n",
			parsedCmd.Command, parsedCmd.Pattern, parsedCmd.Filters)

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
	fmt.Fprintf(os.Stderr, "DEBUG: Command=%s, Days=%d, Chain=%v, Filters=%+v\n",
		cmd.Command, cmd.Days, cmd.Chain, cmd.Filters)

	var results []scanner.Metadata

	results = executeInitialCommand(cmd, s)
	fmt.Fprintf(os.Stderr, "DEBUG: After initial command: %d results\n", len(results))
	for _, r := range results[:min(3, len(results))] {
		fmt.Fprintf(os.Stderr, "DEBUG: Sample - %s (%s)\n", r.Name, r.Extension)
	}

	if len(cmd.Chain) > 0 {
		for _, chainStep := range cmd.Chain {
			results = executeChainStep(chainStep, results, s)
			fmt.Fprintf(os.Stderr, "DEBUG: After chain step '%s': %d results\n", chainStep, len(results))
			for _, r := range results[:min(3, len(results))] {
				fmt.Fprintf(os.Stderr, "DEBUG: Sample - %s (%s)\n", r.Name, r.Extension)
			}
		}
	}
	return results
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func executeInitialCommand(cmd *ai.CommandParser, s *scanner.Scanner) []scanner.Metadata {
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

		if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
			fmt.Fprintf(os.Stderr, "DEBUG: Applying extension filter: '%s'\n", ext)
			fmt.Fprintf(os.Stderr, "DEBUG: Before filter: %d results\n", len(results))
			fmt.Fprintf(os.Stderr, "DEBUG: Sample extensions: %s, %s, %s\n",
				results[0].Extension, results[1].Extension, results[2].Extension)
			results = filterByExtension(results, ext)
			fmt.Fprintf(os.Stderr, "DEBUG: After filter: %d results\n", len(results))
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
		if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
			results = filterByExtension(results, ext)
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
		if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
			results = filterByExtension(results, ext)
		}
	case "search":
		opts := scanner.SearchOptions{
			Pattern:       cmd.Pattern,
			UseRegex:      cmd.Flags["regex"],
			CaseSensitive: cmd.Flags["case_sensitive"],
		}

		fmt.Fprintf(os.Stderr, "DEBUG: Search - Pattern=%s, Regex=%v, CaseSensitive=%v\n",
			opts.Pattern, opts.UseRegex, opts.CaseSensitive)

		if ext, ok := cmd.Filters["extension"]; ok && ext != "" {
			opts.Extensions = strings.Split(ext, ",")
			fmt.Fprintf(os.Stderr, "DEBUG: Search - Extensions=%v\n", opts.Extensions)
		}

		searchResults := s.Search(opts)

		fmt.Fprintf(os.Stderr, "DEBUG: Search found %d files\n", len(searchResults))

		for _, sr := range searchResults {
			for _, item := range s.Results {
				if item.Location == sr.Location {
					results = append(results, item)
					break
				}
			}
		}

	}

	return results
}

func executeChainStep(step string, currentResults []scanner.Metadata, s *scanner.Scanner) []scanner.Metadata {
	parts := strings.Split(step, ":")
	command := parts[0]
	param := ""
	if len(parts) > 1 {
		param = parts[1]
	}

	switch command {
	case "largest":
		// Get N largest from current results
		n := 1
		if param != "" {
			n, _ = strconv.Atoi(param)
		}

		// Sort by size
		sort.Slice(currentResults, func(i, j int) bool {
			si := currentResults[i].Size
			if currentResults[i].IsDir {
				si = currentResults[i].DirSize
			}
			sj := currentResults[j].Size
			if currentResults[j].IsDir {
				sj = currentResults[j].DirSize
			}
			return si > sj
		})

		if n > len(currentResults) {
			n = len(currentResults)
		}
		return currentResults[:n]

	case "smallest":
		// Get N smallest from current results
		n := 1
		if param != "" {
			n, _ = strconv.Atoi(param)
		}

		sort.Slice(currentResults, func(i, j int) bool {
			si := currentResults[i].Size
			if currentResults[i].IsDir {
				si = currentResults[i].DirSize
			}
			sj := currentResults[j].Size
			if currentResults[j].IsDir {
				sj = currentResults[j].DirSize
			}
			return si < sj
		})

		if n > len(currentResults) {
			n = len(currentResults)
		}
		return currentResults[:n]

	case "recent":
		// Filter by days from current results
		days := 7
		if param != "" {
			days, _ = strconv.Atoi(param)
		}

		filterDate := time.Now().AddDate(0, 0, -days)
		filtered := []scanner.Metadata{}
		for _, item := range currentResults {
			if item.ModificationTime.After(filterDate) {
				filtered = append(filtered, item)
			}
		}
		return filtered

	case "oldest":
		// Filter by days from current results
		days := 30
		if param != "" {
			days, _ = strconv.Atoi(param)
		}

		filterDate := time.Now().AddDate(0, 0, -days)
		filtered := []scanner.Metadata{}
		for _, item := range currentResults {
			if item.ModificationTime.Before(filterDate) {
				filtered = append(filtered, item)
			}
		}
		return filtered

	case "search":
		// Search within current results only
		if param == "" {
			return currentResults
		}

		// Create a temporary scanner with only current results
		tempScanner := &scanner.Scanner{
			Results: currentResults,
		}

		opts := scanner.SearchOptions{
			Pattern: param,
		}

		searchResults := tempScanner.Search(opts)

		// Convert back to Metadata
		filtered := []scanner.Metadata{}
		for _, sr := range searchResults {
			for _, item := range currentResults {
				if item.Location == sr.Location {
					filtered = append(filtered, item)
					break
				}
			}
		}
		return filtered

	case "extension":
		// Filter by extension
		filtered := []scanner.Metadata{}
		for _, item := range currentResults {
			if item.Extension == param {
				filtered = append(filtered, item)
			}
		}
		return filtered
	}

	return currentResults
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

// file Actions - second API layer
func performFileAction(filepath, actionType string, customQuery string) tea.Cmd {
	return func() tea.Msg {
		var content string
		var err error

		// Handle different file types
		if utils.IsPDF(filepath) {
			content, err = utils.ExtractPDFText(filepath)
			if err != nil {
				return fileActionCompleteMsg{err: fmt.Errorf("could not read PDF: %v", err)}
			}
		} else if utils.IsTextFile(filepath) {
			contentBytes, readErr := os.ReadFile(filepath)
			if readErr != nil {
				return fileActionCompleteMsg{err: fmt.Errorf("could not read file: %v", readErr)}
			}
			content = string(contentBytes)
		} else {
			return fileActionCompleteMsg{err: fmt.Errorf("unsupported file type - only text and PDF files are supported")}
		}

		// Check if content is empty
		if strings.TrimSpace(content) == "" {
			return fileActionCompleteMsg{err: fmt.Errorf("file appears to be empty or unreadable")}
		}

		// Truncate if too large (optional - keep for now to avoid huge costs)
		const maxChars = 100000 // ~100KB of text
		if len(content) > maxChars {
			content = content[:maxChars] + "\n\n[Content truncated due to length...]"
		}

		// Build prompt based on action
		var prompt string
		switch actionType {
		case "summarize":
			prompt = fmt.Sprintf(`Provide a clear, concise summary of this document in 3-5 sentences. Document:%s`, content)
		case "keypoints":
			prompt = fmt.Sprintf(`Extract the main points from this document. Format as a simple numbered list (1., 2., etc.). Document: %s`, content)
		case "question":
			prompt = fmt.Sprintf("Based on this document: %s\n\nQuestion: %s", content, customQuery)
		case "findinfo":
			prompt = fmt.Sprintf(`Describe the main topics and important details in this document in plain text. Be specific and organized. Document: %s`, content)
		}

		// Call AI
		godotenv.Load()
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fileActionCompleteMsg{err: fmt.Errorf("GEMINI_API_KEY not found")}
		}

		aiClient, err := ai.NewClient(apiKey)
		if err != nil {
			return fileActionCompleteMsg{err: err}
		}

		result, err := aiClient.ProcessDocument(prompt)
		if err != nil {
			return fileActionCompleteMsg{err: err}
		}

		return fileActionCompleteMsg{result: result}
	}
}
