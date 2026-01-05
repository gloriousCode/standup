package standup

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Entry represents a single daily standup report.
type Entry struct {
	DidToday string
	Issues   string
}

const none = "(none)"

// noteIconPrefix is an icon shown at the start of each wrapped note line.
const noteIconPrefix = "🌞"

// Prompt asks the user standup questions, reading answers from r and writing prompts to w.
//
// Input format:
//   - For each question you can type a single line answer.
//   - If you want a multi-line answer, type your text across multiple lines and finish with a single dot (.) on its own line.
func Prompt(r io.Reader, w io.Writer) (Entry, error) {
	br := bufio.NewReader(r)

	did, err := promptOne(br, w, "What did you do today?")
	if err != nil {
		return Entry{}, err
	}

	issues, err := promptOne(br, w, "Any issues you need help with?")
	if err != nil {
		return Entry{}, err
	}

	return Entry{DidToday: did, Issues: issues}, nil
}

func promptOne(br *bufio.Reader, w io.Writer, question string) (string, error) {
	if _, err := fmt.Fprintf(w, "%s\n", question); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintln(w, "(enter a single line, or multiple lines ending with a '.' line)"); err != nil {
		return "", err
	}
	if _, err := fmt.Fprint(w, "> "); err != nil {
		return "", err
	}

	var lines []string
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			// EOF: accept what we have (if any), otherwise propagate.
			line = strings.TrimRight(line, "\r\n")
			if line != "" {
				lines = append(lines, line)
			}
			if len(lines) == 0 {
				return "", err
			}
			break
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "." {
			break
		}
		lines = append(lines, line)

		// single-line shortcut: if the user entered one line and then immediately wants to move on,
		// they can just press Enter on next prompt; we keep reading until '.' or EOF.
		if _, err := fmt.Fprint(w, "> "); err != nil {
			return "", err
		}
	}

	ans := strings.TrimSpace(strings.Join(lines, "\n"))
	if ans == "" {
		return none, nil
	}
	return ans, nil
}

// RenderTable creates a pretty table that contains the entry.
// It returns the rendered string so callers/tests can control where it gets printed.
func RenderTable(e Entry) string {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"Thing", "Notes"})

	// Wrap notes so tables stay readable in narrow terminals.
	// This also makes output deterministic for tests.
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Thing"},
		{Name: "Notes", WidthMax: 60, WidthMaxEnforcer: text.WrapSoft},
	})

	tw.AppendRow(table.Row{"What I did today", formatNotes(e.DidToday)})
	tw.AppendRow(table.Row{"Issues / help needed", formatNotes(e.Issues)})
	return tw.Render()
}

func formatNotes(s string) string {
	// Preserve explicit newlines from multi-line answers, but wrap each paragraph.
	paragraphs := strings.Split(s, "\n")
	var out []string
	for _, p := range paragraphs {
		p = strings.TrimRight(p, "\r")
		if strings.TrimSpace(p) == "" {
			out = append(out, "")
			continue
		}
		wrapped := text.WrapSoft(p, 56) // 60 minus prefix + a little breathing room
		for _, line := range strings.Split(wrapped, "\n") {
			out = append(out, noteIconPrefix+" "+line)
		}
	}
	return strings.Join(out, "\n")
}

// RenderMarkdown creates a markdown table (like hi.md) that contains the entry.
// It uses <br/> for linebreaks so multi-line answers stay within a single markdown table cell.
// The date is included above the table.
func RenderMarkdown(dateDDMMYYYY string, e Entry) string {
	headers := []string{"THING", "NOTES"}
	sep := []string{"----------------------", "-------------------------------------------------------------------------------------------------------------------------------------------------------------"}

	thingWidth := len(headers[0])
	notesWidth := len(headers[1])
	rows := [][2]string{
		{"What I did today", formatNotesMarkdown(e.DidToday)},
		{"Issues / help needed", formatNotesMarkdown(e.Issues)},
	}

	for _, r := range rows {
		if l := len(r[0]); l > thingWidth {
			thingWidth = l
		}
		if l := len(r[1]); l > notesWidth {
			notesWidth = l
		}
	}
	if thingWidth < len(sep[0]) {
		thingWidth = len(sep[0])
	}
	if notesWidth < len(sep[1]) {
		notesWidth = len(sep[1])
	}

	var b strings.Builder
	b.WriteString("\n")
	if strings.TrimSpace(dateDDMMYYYY) != "" {
		b.WriteString("Date: " + dateDDMMYYYY + "\n\n")
	}
	b.WriteString(fmt.Sprintf("| %-*s | %-*s |\n", thingWidth, headers[0], notesWidth, headers[1]))
	b.WriteString(fmt.Sprintf("|%s|%s|\n", strings.Repeat("-", thingWidth+2), strings.Repeat("-", notesWidth+2)))
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("| %-*s | %-*s |\n", thingWidth, r[0], notesWidth, r[1]))
	}
	b.WriteString("\n")
	return b.String()
}

func formatNotesMarkdown(s string) string {
	// Preserve explicit newlines from multi-line answers.
	// In markdown tables, raw newlines break the table, so we convert them to <br/>.
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimRight(l, "\r")
		if strings.TrimSpace(l) == "" {
			continue
		}
		// Escape literal pipes as they break markdown tables.
		l = strings.ReplaceAll(l, "|", "\\|")
		out = append(out, noteIconPrefix+" "+strings.TrimSpace(l))
	}
	if len(out) == 0 {
		return noteIconPrefix + " " + none
	}
	return strings.Join(out, " <br/>")
}
