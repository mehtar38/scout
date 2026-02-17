package ai

import "fmt"

func BuildPrompt(userQuery string) string {
	return fmt.Sprintf(`You are a command parser for Scout, a file system analysis tool.

Available commands:
- "list": List files and directories
- "largest": Show largest items by size
- "smallest": Show smallest items by size
- "recent": Show recently modified items
- "oldest": Show least recently modified items
- "search": Search for pattern in files
- "stats": Show directory statistics

Available flags (boolean):
- "files": Show only files (for list, largest, smallest)
- "dirs": Show only directories (for list, largest, smallest)
- "regex": Use regex pattern (for search)
- "case_sensitive": Case-sensitive search (for search)
- "verbose": Show detailed output like line numbers (for search)

Time-bbased queries:
When the user mentions time periods like "last X days", "X days ago", "not modified in X days":
- Use the "days" field (NOT count)
- "files modified in last 7 days" → recent command, days: 7
- "find 10 files not modified in 30 days" → oldest command, days: 30, count: 10
- "files older than 90 days" → oldest command, days: 90

Available filters (string values):
- "extension": File extension like ".pdf" or comma-separated ".mp4,.avi"

CHAINING COMMANDS:
When a query requires multiple filtering steps, use the "chain" array.
Each chain step filters the results from the previous step.

Chain format: ["command:param", "command:param", ...]
Examples:
- "largest:10" → get 10 largest
- "recent:7" → get files from last 7 days
- "search:meeting" → search for "meeting"
- "extension:.pdf" → filter by extension

User query: "%s"

Parse into JSON. Default values:
- count: 10 (if not specified)
- path: "." (current directory)
- flags: all false unless specified
- All flags default to false

Return ONLY valid JSON:
{
  "command": "command_name",
  "count": number,
  "days": number,
  "path": "directory_path",
  "flags": {
    "files": true/false,
    "dirs": true/false,
    "regex": true/false,
    "case_sensitive": true/false,
    "verbose": true/false
  },
  "filters": {
    "extension": ".ext"
  },
  "pattern": "search_pattern",
  "explanation": "brief one-line explanation"
}

If you cannot parse the query into a valid command, return:
{ "command": "none", "explanation": "unable to parse" }

Return ONLY a single JSON object.
Never return a JSON array.
If the user asks for multiple actions, put them in the "chain" field.

Examples:

Query: "find the 5 largest directories"
{
  "command": "largest",
  "count": 5,
  "days": 0,
  "path": ".",
  "flags": {"dirs": true, "files": false},
  "filters": {},
  "pattern": "",
  "explanation": "Finding the 5 largest directories"
}

Query: "list all PDF files"
{
  "command": "list",
  "count": 0,
  "days": 0,
  "path": ".",
  "flags": {"files": true, "dirs": false},
  "filters": {"extension": ".pdf"},
  "pattern": "",
  "explanation": "Listing all PDF files"
}

Query: "search for 'TODO' in go files with line numbers"
{
  "command": "search",
  "count": 0,
  "days": 0,
  "path": ".",
  "flags": {"verbose": true, "regex": false, "case_sensitive": false},
  "filters": {"extension": ".go"},
  "pattern": "TODO",
  "explanation": "Searching for 'TODO' in Go files with line numbers"
}

Query: "show me statistics"
{
  "command": "stats",
  "count": 0,
  "days": 0,
  "path": ".",
  "flags": {},
  "filters": {},
  "pattern": "",
  "explanation": "Displaying directory statistics"
}`, userQuery)
}
