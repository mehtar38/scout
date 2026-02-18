# 🔍 Scout - AI-Powered File System Analysis Tool

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Scout is a powerful, AI-enhanced command-line file system analysis tool with an interactive Terminal User Interface (TUI). Navigate, search, analyze, and interact with your files using natural language queries and intelligent automation.

## ✨ Features

### 🎯 Core Capabilities
- **Fast File Scanning**: Concurrent directory traversal with Go routines
- **Natural Language Queries**: Ask questions like "find the 10 largest PDFs from last week"
- **Interactive TUI**: Beautiful terminal interface with mouse-free navigation
- **Content Search**: Regex-powered search across file contents
- **AI-Powered Analysis**: Summarize documents, extract key points, and answer questions about files
- **Multiple File Formats**: Text files, PDFs, code files, and more
- **Command Chaining**: Complex queries like "find PDFs with 'meeting' and show the 5 largest"

### 🖥️ Dual Interface
- **CLI Mode**: Fast, scriptable commands for automation
- **TUI Mode**: Interactive dashboard for exploration and discovery

### 🤖 AI Integration
- Natural language command parsing
- Document summarization
- Key point extraction
- Question answering about file contents
- Intelligent query chaining

## 📸 Screenshots

> Add your screenshots here:
> - TUI Dashboard
> - Browser with AI results
> - Search interface
> - AI action menu

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Gemini API key (for AI features)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/scout.git
   cd scout
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up API key**
   ```bash
   # Create .env file
   echo "GEMINI_API_KEY=your-api-key-here" > .env
   ```

4. **Install globally**
   ```bash
   go install
   ```

5. **Run Scout**
   ```bash
   scout --help
   ```

## 📖 Usage

### CLI Commands

#### Basic File Operations

```bash
# List files
scout list [path]
scout list --files          # Files only
scout list --dirs           # Directories only

# Find largest/smallest
scout largest 10            # 10 largest files
scout largest 5 /path       # In specific directory
scout smallest 10 --dirs    # Smallest directories

# Time-based queries
scout recent 20             # 20 most recently modified
scout recent --days 7       # Files from last 7 days
scout oldest --days 30      # Not modified in 30 days

# Search content
scout search "TODO"
scout search "error" --regex
scout search "pattern" --case-sensitive --verbose

# Statistics
scout stats [path]
```

#### AI-Powered Queries

```bash
# Natural language file discovery
scout ai "find the 10 largest PDF files"
scout ai "show me Go files modified this week"
scout ai "find documents with 'meeting' in them"
scout ai "what are the oldest files not touched in 90 days"

# Complex chained queries
scout ai "find 5 largest PDFs from last month"
scout ai "search for TODO in Go files and show the 3 largest"
```

### Interactive TUI

Launch the Terminal UI:

```bash
scout tui [path]
```

#### TUI Navigation

**Global Controls:**
- `1-4` or `Tab` - Switch between tabs
- `r` / `F5` - Refresh/rescan directory
- `q` / `Ctrl+C` - Quit
- `↑↓` or `jk` - Navigate (Vim-style supported)

**Dashboard Tab (1):**
- View statistics overview
- File/directory breakdown
- Size analysis
- Top extensions

**Browser Tab (2):**
- `s` - Cycle sort mode (name/size/time/extension)
- `S` - Reverse sort order
- `f` - Filter (all/files/dirs)
- `e` - Cycle through extensions
- `o` / `Enter` - Open file location in system explorer
- `c` - Clear filters and reset
- `a` - AI actions (on AI results)

**Search Tab (3):**
- `/` or `Enter` - Start search
- `r` - Toggle regex mode
- `c` - Toggle case sensitivity
- `o` - Open selected file
- `Esc` - Cancel input

**AI Tab (4):**
- `/` or `Enter` - Ask AI question
- `v` - View results in Browser tab
- `Esc` - Cancel input

#### AI File Actions

After running an AI query and viewing results in Browser:

1. Navigate to any file
2. Press `a` to open action menu
3. Choose action:
   - **Summarize** - Get 3-5 sentence summary
   - **Key Points** - Extract main points as numbered list
   - **Ask Question** - Ask specific questions about the file
   - **Find Information** - Get detailed analysis
4. View AI-generated result
5. Press `Esc` to close

## 🏗️ Architecture

### Project Structure

```
scout/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── ai/                     # AI integration
│   │   ├── geminiClient.go     # Gemini API client
│   │   └── prompt.go           # Query parsing prompts
|   |   |__ types.go
│   ├── commands/               # CLI commands
│   │   ├── root.go             # Root command setup
│   │   ├── ai.go               # AI command handler
│   │   ├── list.go             # List command
│   │   ├── largest.go          # Largest command
│   │   ├── smallest.go         # Smallest command
│   │   ├── recent.go           # Recent command
│   │   ├── oldest.go           # Oldest command
│   │   ├── search.go           # Search command
│   │   ├── stats.go            # Stats command
│   │   └── tui.go              # TUI launcher
│   ├── scanner/                # File system operations
│   │   ├── scanner.go          # Core scanning logic
│   │   └── search.go           # Content search
|   |   |__ sequentialSearch.go # Content seatch without concurrency (commented out - only written for leanring purposes)
│   ├── tui/                    # Terminal UI (MVC pattern)
│   │   ├── model.go            # State management
│   │   ├── update.go           # Event handling
│   │   └── view.go             # Rendering
│   └── utils/                  # Utilities
│       ├── pagination.go            # CLI pagination
│       └── pdf.go              # PDF text extraction
├── go.mod
├── go.sum
├── .env                        # API keys (not in repo)
└── README.md
```

### Technology Stack

**Core:**
- **Language**: Go 1.21+
- **Concurrency**: Go routines for fast scanning and searching

**CLI Framework:**
- **[Cobra](https://github.com/spf13/cobra)**: Command-line interface
- **[godotenv](https://github.com/joho/godotenv)**: Environment configuration

**TUI Framework:**
- **[BubbleTea](https://github.com/charmbracelet/bubbletea)**: The Elm Architecture for terminals
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Style definitions and layout

**AI Integration:**
- **[Google Gemini API](https://ai.google.dev/)**: Natural language processing
- **[genai-go](https://github.com/google/generative-ai-go)**: Official Gemini SDK

**File Processing:**
- **[pdf](https://github.com/ledongthuc/pdf)**: PDF text extraction
- **Standard Library**: File I/O, regex, time handling

### Design Patterns

- **MVC Architecture** (TUI): Clean separation of Model, Update (Controller), View
- **Command Pattern** (CLI): Each command is self-contained and composable
- **Strategy Pattern** (Search): Different search strategies (regex, case-sensitive, etc.)
- **Factory Pattern** (Scanner): Scanner creation with validation
- **Chain of Responsibility** (AI Queries): Sequential query processing

## 🔧 Development

### Building from Source

```bash
# Development build
go build -o scout cmd/main.go

# Run without installing
go run cmd/main.go <command>

# Run tests
go test ./...

# Run with debugging
go run cmd/main.go tui 2> debug.log
```

### Project Goals

Scout was built to:
- Learn Go through practical application
- Explore concurrent programming patterns
- Build production-quality CLI tools
- Integrate modern AI capabilities
- Create an intuitive TUI with BubbleTea
  
## 📚 Key Features Explained

### Command Chaining

Scout can chain multiple operations:

```bash
scout ai "find PDFs with 'report' and show 5 largest from last month"
```

This breaks down into:
1. Filter by extension (`.pdf`)
2. Search content for "report"
3. Filter by time (last 30 days)
4. Sort by size
5. Take top 5

### Concurrent Scanning

Search operation uses worker pools:
- Leverages multiple CPU cores
- Concurrent directory walking
- Non-blocking I/O operations
- Efficient for large directories

### Smart Query Parsing

The AI parser understands:
- **Time expressions**: "last week", "3 days ago", "past month"
- **Size queries**: "largest", "smallest", "biggest 10"
- **Content search**: "files containing X"
- **Filters**: "PDF files", "Go files", "directories"
- **Combinations**: All of the above together

## 🎨 TUI Design Philosophy

- **Keyboard-first**: No mouse required
- **Vim-style navigation**: `hjkl` supported everywhere
- **Progressive disclosure**: Show only what's needed
- **Visual hierarchy**: Color and spacing guide attention
- **Responsive**: Adapts to terminal size
- **Fast**: Optimized rendering, no flicker

## 🔒 Privacy & Security

- **Local processing**: All file scanning happens on your machine
- **API calls**: Only when using AI features
- **No data collection**: Scout doesn't track or store anything
- **API key security**: Stored locally in `.env` file
- **Read-only by default**: TUI is view-only (for now)

## 🗺️ Roadmap

### Current Features (v1.0)
- ✅ CLI commands for all file operations
- ✅ Interactive TUI with 4 tabs
- ✅ AI-powered natural language queries
- ✅ Command chaining
- ✅ PDF support for AI actions
- ✅ Content search with regex

### Planned Features (v2.0)
- [ ] File operations (move, delete, copy) in TUI
- [ ] Multi-file selection
- [ ] Export results to CSV/JSON
- [ ] File comparison 
- [ ] Configuration file support
- [ ] Bookmarks/favorites
- [ ] File tagging system
- [ ] Integration with cloud storage

### Long-term Vision
- [ ] Web interface option
- [ ] Advanced AI agents
- [ ] Natural language file organization

## 🤝 Contributing

Contributions are welcome! This project was built as a learning experience, and I'd love to see how others might extend it.

### Areas for Contribution

- Additional AI actions
- More file format support
- Performance optimizations
- UI/UX improvements
- Documentation
- Test coverage
- Bug fixes

### Development Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Follow Go best practices and project architecture
5. Test thoroughly
6. Commit with clear messages (`git commit -m 'Add amazing feature'`)
7. Push to your branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Charm](https://charm.sh/)** - For the amazing BubbleTea and Lipgloss libraries
- **[Google AI](https://ai.google.dev/)** - For the Gemini API
- **[Cobra](https://github.com/spf13/cobra)** - For the excellent CLI framework
- **Go Community** - For comprehensive documentation and support

## 📧 Contact

**Project Link**: [https://github.com/yourusername/scout](https://github.com/yourusername/scout)

---

<div align="center">

⭐ Star this repo if you find it useful!

</div>
