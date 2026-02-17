package tui

import (
	"fmt"
	"scout/internal/scanner"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#04B575")
	accentColor    = lipgloss.Color("#F780E2")
	textColor      = lipgloss.Color("#FAFAFA")
	dimColor       = lipgloss.Color("#626262")
	errorColor     = lipgloss.Color("#FF5F87")

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Padding(0, 2)

	// Header styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Content styles
	listItemStyle = lipgloss.NewStyle().
			Foreground(textColor)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Background(primaryColor).
				Bold(true)

	dimTextStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Border styles
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Status bar style
	statusBarStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(dimColor).
			Padding(0, 1)

	// Input styles
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(0, 1)
)

// View renders the entire TUI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Header with title
	b.WriteString(titleStyle.Render("Scout"))
	b.WriteString("\n\n")

	// Tab bar
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Main content area
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	} else if m.scanning {
		b.WriteString(dimTextStyle.Render("Scanning directory..."))
	} else if !m.scanned {
		b.WriteString(dimTextStyle.Render("⏳ Scout is scanning your files..."))
	} else {
		b.WriteString(m.renderActiveTab())
	}

	// Status bar
	b.WriteString("\n\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderTabs renders the tab navigation bar
func (m Model) renderTabs() string {
	tabs := []string{"Dashboard", "Browser", "Search", "AI"}
	var rendered []string

	for i, tab := range tabs {
		if Tab(i) == m.activeTab {
			rendered = append(rendered, activeTabStyle.Render(fmt.Sprintf("%d. %s", i+1, tab)))
		} else {
			rendered = append(rendered, inactiveTabStyle.Render(fmt.Sprintf("%d. %s", i+1, tab)))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// renderActiveTab renders the content of the active tab
func (m Model) renderActiveTab() string {
	switch m.activeTab {
	case TabDashboard:
		return m.renderDashboard()
	case TabBrowser:
		return m.renderBrowser()
	case TabSearch:
		return m.renderSearch()
	case TabAI:
		return m.renderAI()
	default:
		return "Unknown tab"
	}
}

// renderDashboard renders the dashboard view with stats
func (m Model) renderDashboard() string {
	var content strings.Builder

	files := m.scanner.GetFiles()
	dirs := m.scanner.GetDirectories()

	// Build ALL content first (don't truncate)
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("System Statistics"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("  Total Items: %d\n", len(files)+len(dirs)))
	content.WriteString(fmt.Sprintf("  Files: %d\n", len(files)))
	content.WriteString(fmt.Sprintf("  Directories: %d\n", len(dirs)))
	content.WriteString("\n")

	// Size statistics
	var totalFileSize, totalDirSize int64
	for _, f := range files {
		totalFileSize += f.Size
	}
	for _, d := range dirs {
		totalDirSize += d.DirSize
	}

	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("Size Statistics"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("  Total Size: %s\n", formatSize(totalFileSize+totalDirSize)))
	content.WriteString(fmt.Sprintf("  Files: %s\n", formatSize(totalFileSize)))
	content.WriteString(fmt.Sprintf("  Directories: %s\n", formatSize(totalDirSize)))
	if len(files) > 0 {
		content.WriteString(fmt.Sprintf("  Avg File Size: %s\n", formatSize(totalFileSize/int64(len(files)))))
	}
	if len(dirs) > 0 {
		content.WriteString(fmt.Sprintf("  Avg Dir Size: %s\n", formatSize(totalDirSize/int64(len(dirs)))))
	}
	content.WriteString("\n")

	// Largest and smallest
	if len(files) > 0 {
		largest := m.scanner.GetNLargestFilesBySize(1)
		smallest := m.scanner.GetNSmallestFilesBySize(1)

		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("Extremes"))
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("  Largest: %s (%s)\n", largest[0].Name, formatSize(largest[0].Size)))
		content.WriteString(fmt.Sprintf("  Smallest: %s (%s)\n", smallest[0].Name, formatSize(smallest[0].Size)))
		content.WriteString("\n")
	}

	// Recent activity
	if len(files) > 0 {
		recent := m.scanner.GetNRecentlyModFiles(1)
		oldest := m.scanner.GetNLeastModFiles(1)

		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("Modification Times"))
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("  Most Recent: %s (%s)\n", recent[0].Name, recent[0].ModificationTime.Format("2006-01-02 15:04")))
		content.WriteString(fmt.Sprintf("  Oldest: %s (%s)\n", oldest[0].Name, oldest[0].ModificationTime.Format("2006-01-02 15:04")))
		content.WriteString("\n")
	}

	// File type breakdown
	extCounts := make(map[string]int)
	for _, f := range files {
		if f.Extension != "" {
			extCounts[f.Extension]++
		}
	}

	if len(extCounts) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("Extensions"))
		content.WriteString("\n\n")

		type extCount struct {
			ext   string
			count int
		}
		var exts []extCount
		for ext, count := range extCounts {
			exts = append(exts, extCount{ext, count})
		}
		sort.Slice(exts, func(i, j int) bool {
			return exts[i].count > exts[j].count
		})

		for i, ec := range exts {
			if i >= 10 { // Show top 10
				break
			}
			content.WriteString(fmt.Sprintf("  %s: %d files\n", ec.ext, ec.count))
		}
	}

	// Now apply scrolling to the content
	allLines := strings.Split(content.String(), "\n")
	visibleHeight := m.height - 12 // Account for header/footer

	start := m.scrollOffset
	end := start + visibleHeight
	if end > len(allLines) {
		end = len(allLines)
	}

	visibleContent := strings.Join(allLines[start:end], "\n")

	if len(allLines) > visibleHeight {
		visibleContent += "\n" + dimTextStyle.Render(fmt.Sprintf("Showing lines %d-%d of %d", start+1, end, len(allLines)))
	}

	return borderStyle.Render(visibleContent)
}

// renders the file browser view
func (m Model) renderBrowser() string {
	var b strings.Builder

	// Header with controls
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("File Browser"))
	b.WriteString("\n\n")
	b.WriteString(dimTextStyle.Render(fmt.Sprintf("Sort: %s %s | Filter: %s | Extension: %s",
		m.getSortModeName(),
		m.getSortDirection(),
		m.getFilterModeName(),
		m.getExtensionFilter())))
	b.WriteString("\n")
	b.WriteString(dimTextStyle.Render("s: sort | S: reverse | f: filter | e: ext. filter | o: open  | c: reset"))
	b.WriteString("\n\n")

	// Get and display items
	items := m.getBrowserItems()

	if len(items) == 0 {
		b.WriteString(dimTextStyle.Render("No items to display"))
		return borderStyle.Render(b.String())
	}

	// Calculate visible range
	visibleHeight := m.height - 15 // Account for headers, footer, padding
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	start := m.scrollOffset
	end := start + visibleHeight
	if end > len(items) {
		end = len(items)
	}

	// Render items
	for i := start; i < end; i++ {
		item := items[i]
		icon := "📄"
		if item.IsDir {
			icon = "📁"
		}

		sizeStr := formatSize(item.Size)
		if item.IsDir {
			sizeStr = formatSize(item.DirSize)
		}

		line := fmt.Sprintf("%s %-40s %10s  %s",
			icon,
			truncate(item.Name, 40),
			sizeStr,
			item.ModificationTime.Format("2006-01-02 15:04"))

		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render(line))
		} else {
			b.WriteString(listItemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Show scroll indicator
	if len(items) > visibleHeight {
		b.WriteString("\n")
		b.WriteString(dimTextStyle.Render(fmt.Sprintf("Showing %d-%d of %d items", start+1, end, len(items))))
	}

	return borderStyle.Render(b.String())
}

// renders the search view
func (m Model) renderSearch() string {
	var b strings.Builder

	// Header
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("Search"))
	b.WriteString("\n\n")

	// Search input
	if m.searchActive {
		b.WriteString("Enter search pattern:\n")
		b.WriteString(inputBoxStyle.Render("> " + m.searchInput + "█"))
		b.WriteString("\n")
	} else {
		b.WriteString("Press Enter or / to start searching\n")
		if m.searchPattern != "" {
			b.WriteString(fmt.Sprintf("Current pattern: %s\n", m.searchPattern))
		}
	}

	// Options
	b.WriteString("\n")
	regexStatus := "off"
	if m.searchRegex {
		regexStatus = "ON"
	}
	caseStatus := "off"
	if m.searchCaseSensitive {
		caseStatus = "ON"
	}
	b.WriteString(dimTextStyle.Render(fmt.Sprintf("r: regex [%s] | c: case-sensitive [%s]| o: open", regexStatus, caseStatus)))
	b.WriteString("\n\n")

	if m.searchLoading {
		b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render("⏳ Scout's on it..."))
		return borderStyle.Render(b.String())
	}

	// Results
	if len(m.searchResults) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Found matches in %d files:", len(m.searchResults))))
		b.WriteString("\n\n")

		headerLines := strings.Count(b.String(), "\n")
		visibleHeight := m.height - headerLines - 10 // Account for borders
		if visibleHeight < 3 {
			visibleHeight = 3
		}

		start := m.scrollOffset
		end := start + visibleHeight
		if end > len(m.searchResults) {
			end = len(m.searchResults)
		}

		for i := start; i < end; i++ {
			result := m.searchResults[i]
			line := fmt.Sprintf("  %s: %d matches", result.Location, result.Matches)

			if i == m.cursor {
				b.WriteString(selectedItemStyle.Render(line))
			} else {
				b.WriteString(listItemStyle.Render(line))
			}
			b.WriteString("\n")
		}

		// Show scroll indicator
		if len(m.searchResults) > visibleHeight {
			b.WriteString("\n")
			b.WriteString(dimTextStyle.Render(fmt.Sprintf("Showing %d-%d of %d results", start+1, end, len(m.searchResults))))
		}

	} else if m.searchPattern != "" && !m.searchActive && !m.searchLoading {
		b.WriteString(dimTextStyle.Render("No matches found"))
	}

	return borderStyle.Render(b.String())
}

// renderAI renders the AI query view
func (m Model) renderAI() string {
	var b strings.Builder

	// Header
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("Scout AI Assistant"))
	b.WriteString("\n\n")

	// Input box (highest priority - show when active)
	if m.aiActive {
		b.WriteString("Enter your query:\n")
		b.WriteString(inputBoxStyle.Render("> " + m.aiInput + "█"))
		b.WriteString("\n\n")
		b.WriteString(dimTextStyle.Render("Example: \"find all PDF files larger than 5MB\""))
		return borderStyle.Render(b.String())
	}

	// Loading state (second priority - show while processing)
	if m.aiLoading {
		b.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render("⏳ Processing your query..."))
		b.WriteString("\n")
		if m.aiQuery != "" {
			b.WriteString(dimTextStyle.Render(fmt.Sprintf("Query: %s", m.aiQuery)))
		}
		return borderStyle.Render(b.String())
	}

	// Normal view (show results or prompts)
	b.WriteString("Press Enter or / to ask a question\n\n")

	// Show last query and response
	if m.aiQuery != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Last Query:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.aiQuery))

		if m.aiResponse != "" {
			b.WriteString(lipgloss.NewStyle().Bold(true).Render("Response:"))
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("  %s\n\n", m.aiResponse))
		}

		if len(m.aiResults) > 0 {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render(fmt.Sprintf("✓ Found %d results", len(m.aiResults))))
			b.WriteString("\n")
			b.WriteString(dimTextStyle.Render("Press 'v' to view results in Browser tab"))
			b.WriteString("\n\n")

			// Show preview of first few results
			b.WriteString("Preview:\n")
			previewCount := 3
			if len(m.aiResults) < previewCount {
				previewCount = len(m.aiResults)
			}
			for i := 0; i < previewCount; i++ {
				item := m.aiResults[i]
				icon := "📄"
				if item.IsDir {
					icon = "📁"
				}
				b.WriteString(fmt.Sprintf("  %s %s (%s)\n", icon, item.Name, formatSize(item.Size)))
			}
			if len(m.aiResults) > previewCount {
				b.WriteString(dimTextStyle.Render(fmt.Sprintf("  ... and %d more", len(m.aiResults)-previewCount)))
			}
		}
	} else {
		b.WriteString(dimTextStyle.Render("Navigate through with Scout! Try asking:\n"))
		b.WriteString(dimTextStyle.Render("\n• show me the 10 largest files"))
		b.WriteString(dimTextStyle.Render("\n• find all images modified this week"))
		b.WriteString(dimTextStyle.Render("\n• list all Go files"))
	}

	return borderStyle.Render(b.String())
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	left := fmt.Sprintf("📂 %s", m.path)

	middle := ""
	if m.scanned {
		items := len(m.scanner.Results)

		// Show different counts based on active tab
		if m.activeTab == TabBrowser && len(m.aiResults) > 0 {
			middle = fmt.Sprintf("AI Results: %d items", len(m.aiResults))
		} else {
			middle = fmt.Sprintf("%d items", items)
		}
	}

	right := "q: quit | r: refresh | tab: switch | ↑↓: navigate"

	// Calculate spacing
	totalWidth := m.width
	leftLen := len(left)
	middleLen := len(middle)
	rightLen := len(right)

	spacing := totalWidth - leftLen - middleLen - rightLen
	if spacing < 2 {
		spacing = 2
	}
	leftPad := spacing / 2
	rightPad := spacing - leftPad

	status := left + strings.Repeat(" ", leftPad) + middle + strings.Repeat(" ", rightPad) + right

	// Purple border style instead of gray background
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Foreground(textColor).
		Width(m.width-4).
		Padding(0, 1).
		Render(status)
}

// returns the items to display in browser based on filters and sorts
func (m Model) getBrowserItems() []scanner.Metadata {
	var items []scanner.Metadata

	// If we have AI results to display, show those
	if len(m.aiResults) > 0 && m.activeTab == TabBrowser {
		items = m.aiResults
	} else {
		switch m.filterMode {
		case FilterFiles:
			items = m.scanner.GetFiles()
		case FilterDirs:
			items = m.scanner.GetDirectories()
		default:
			items = m.scanner.Results
		}

		if m.filterExt != "" {
			filtered := []scanner.Metadata{}
			for _, item := range items {
				if item.Extension == m.filterExt {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
	}

	if m.showingAIResults {
		switch m.filterMode {
		case FilterFiles:
			filtered := []scanner.Metadata{}
			for _, item := range items {
				if !item.IsDir {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		case FilterDirs:
			filtered := []scanner.Metadata{}
			for _, item := range items {
				if item.IsDir {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
	}

	if m.filterExt != "" {
		filtered := []scanner.Metadata{}
		for _, item := range items {
			if item.Extension == m.filterExt {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	// Apply sort
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch m.sortMode {
		case SortByName:
			less = items[i].Name < items[j].Name
		case SortBySize:
			si := items[i].Size
			if items[i].IsDir {
				si = items[i].DirSize
			}
			sj := items[j].Size
			if items[j].IsDir {
				sj = items[j].DirSize
			}
			less = si < sj
		case SortByModTime:
			less = items[i].ModificationTime.Before(items[j].ModificationTime)
		case SortByExtension:
			less = items[i].Extension < items[j].Extension
		}

		if m.sortReverse {
			return !less
		}
		return less
	})

	return items
}

func (m Model) getSortModeName() string {
	switch m.sortMode {
	case SortByName:
		return "Name"
	case SortBySize:
		return "Size"
	case SortByModTime:
		return "Modified"
	case SortByExtension:
		return "Extension"
	default:
		return "Unknown"
	}
}

func (m Model) getSortDirection() string {
	if m.sortReverse {
		return "↓"
	}
	return "↑"
}

func (m Model) getFilterModeName() string {
	switch m.filterMode {
	case FilterAll:
		return "All"
	case FilterFiles:
		return "Files"
	case FilterDirs:
		return "Dirs"
	default:
		return "Unknown"
	}
}

func (m Model) getExtensionFilter() string {
	if m.filterExt == "" {
		return "none"
	}
	return m.filterExt
}

func formatSize(bytes int64) string {
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
