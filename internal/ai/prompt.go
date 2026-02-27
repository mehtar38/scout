package ai

import "fmt"

func BuildPrompt(userQuery string) string {
	return fmt.Sprintf(`You are a command parser for Scout, a file system analysis tool.

AVAILABLE COMMANDS:
- "list": List files and directories
- "largest": Show largest items by size
- "smallest": Show smallest items by size
- "recent": Show recently modified items
- "oldest": Show least recently modified items
- "search": Search for pattern in file contents
- "stats": Show directory statistics

AVAILABLE FLAGS (boolean):
- "files": Show only files
- "dirs": Show only directories
- "regex": Use regex pattern (search only)
- "case_sensitive": Case-sensitive search (search only)
- "verbose": Show line numbers (search only)

AVAILABLE FILTERS (string):
- "extension": File extension e.g. ".pdf" or comma-separated ".pdf,.docx"

TIME-BASED QUERIES:
- Use "days" field for time periods, never "count"
- "last X days" / "past X days" → recent, days: X
- "not modified in X days" / "older than X days" → oldest, days: X
- "last week" → days: 7, "last month" → days: 30

CHAINING RULES:
Use "chain" array when query needs multiple steps.
Format: ["command:param"]
- Extension + size query → list with extension filter, chain largest/smallest
- Time + size query → recent/oldest with days, chain largest/smallest
- Extension + time + size → list with extension, chain recent/oldest, chain largest/smallest

RESPONSE FORMAT:
{
  "command": "command_name",
  "count": number,
  "days": number,
  "path": ".",
  "flags": {"files": false, "dirs": false, "regex": false, "case_sensitive": false, "verbose": false},
  "filters": {"extension": ""},
  "pattern": "",
  "chain": [],
  "explanation": "one line explanation"
}

RULES:
- Return ONLY valid JSON, no markdown, no extra text
- Default count: 10
- Default path: "."
- All flags default to false
- If query is unclear: {"command": "none", "explanation": "reason"}

EXAMPLES:

Query: "find 5 largest directories"
{"command":"largest","count":5,"days":0,"path":".","flags":{"dirs":true,"files":false},"filters":{},"pattern":"","chain":[],"explanation":"5 largest directories by size"}

Query: "list all PDF files"
{"command":"list","count":0,"days":0,"path":".","flags":{"files":true,"dirs":false},"filters":{"extension":".pdf"},"pattern":"","chain":[],"explanation":"List all PDF files"}

Query: "find 10 largest PDF and DOCX files"
{"command":"list","count":0,"days":0,"path":".","flags":{"files":true},"filters":{"extension":".pdf,.docx"},"pattern":"","chain":["largest:10"],"explanation":"Filter PDFs and DOCXs then get 10 largest"}

Query: "find largest file from last week"
{"command":"recent","count":0,"days":7,"path":".","flags":{"files":true},"filters":{},"pattern":"","chain":["largest:1"],"explanation":"Get files from last week then find largest"}

Query: "find 4 largest PDFs modified in last 5 days"
{"command":"list","count":0,"days":0,"path":".","flags":{"files":true},"filters":{"extension":".pdf"},"pattern":"","chain":["recent:5","largest:4"],"explanation":"Filter PDFs, keep last 5 days, get 4 largest"}

Query: "search for TODO in go files"
{"command":"search","count":0,"days":0,"path":".","flags":{"files":true},"filters":{"extension":".go"},"pattern":"TODO","chain":[],"explanation":"Search for TODO in Go files"}

Query: "files not modified in 30 days"
{"command":"oldest","count":0,"days":30,"path":".","flags":{"files":true},"filters":{},"pattern":"","chain":[],"explanation":"Files not modified in last 30 days"}

Query: "show statistics"
{"command":"stats","count":0,"days":0,"path":".","flags":{},"filters":{},"pattern":"","chain":[],"explanation":"Directory statistics"}

Query: "find all meeting files"
{
  "command": "search",
  "pattern": "meeting",
  "flags": {"files": true},
  "filters": {},
  "chain": [],
  "explanation": "Search for 'meeting' in file contents"
}

User query: "%s"`, userQuery)
}
