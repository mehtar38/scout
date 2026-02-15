package tui

import (
	"scout/internal/scanner"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Tab int

const (
	TabDashboard Tab = iota //sequence incrementing
	TabBrowser
	TabSearch
	TabAI
)

type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
	SortByModTime
	SortByExtension
)

type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterFiles
	FilterDirs
)

// tui state
type Model struct {
	scanner     *scanner.Scanner
	path        string
	scanned     bool
	scanning    bool
	lastRefresh time.Time

	activeTab    Tab
	cursor       int
	scrollOffset int // For scrolling long lists

	// Browser state
	sortMode    SortMode
	sortReverse bool
	filterMode  FilterMode
	filterExt   string
	showHidden  bool

	// Search state
	searchPattern       string
	searchRegex         bool
	searchCaseSensitive bool
	searchResults       []scanner.SearchResult
	searchInput         string
	searchActive        bool // Whether we're typing in search box

	// AI state
	aiQuery    string
	aiResponse string
	aiInput    string
	aiActive   bool // Whether we're typing in AI box
	aiResults  []scanner.Metadata

	width  int
	height int
	ready  bool

	err error
}

func NewModel(path string) Model {
	return Model{
		path:        path,
		activeTab:   TabDashboard,
		sortMode:    SortByName,
		filterMode:  FilterAll,
		width:       80,
		height:      24,
		scanned:     false,
		scanning:    false,
		lastRefresh: time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return scanDirectory(m.path)
}

// Messages
type scanCompleteMsg struct {
	scanner *scanner.Scanner
	err     error
}

type searchCompleteMsg struct {
	results []scanner.SearchResult
	err     error
}

// Commands
func scanDirectory(path string) tea.Cmd {
	return func() tea.Msg {
		s, err := scanner.New(path)
		if err != nil {
			return scanCompleteMsg{nil, err}
		}

		if err := s.Scan(); err != nil {
			return scanCompleteMsg{nil, err}
		}

		s.ComputeDirSize()

		return scanCompleteMsg{s, nil}
	}
}

func executeSearch(s *scanner.Scanner, pattern string, useRegex, caseSensitive bool) tea.Cmd {
	return func() tea.Msg {
		opts := scanner.SearchOptions{
			Pattern:       pattern,
			UseRegex:      useRegex,
			CaseSensitive: caseSensitive,
		}

		results := s.Search(opts)
		return searchCompleteMsg{results, nil}
	}
}
