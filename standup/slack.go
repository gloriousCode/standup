package standup

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

// RenderSlack creates a Slack-friendly plain-text message.
// It avoids Markdown tables (Slack doesn't render them) and uses wrapped bullet lines.
//
// The output is designed to be copy/pasted into Slack while still looking tidy.
func RenderSlack(dateDDMMYYYY string, e Entry) string {
	var b strings.Builder

	if strings.TrimSpace(dateDDMMYYYY) != "" {
		b.WriteString("Date: " + dateDDMMYYYY + "\n\n")
	}

	b.WriteString("What I did today\n")
	b.WriteString(formatNotesSlack(e.DidToday))
	b.WriteString("\n\n")

	b.WriteString("Issues / help needed\n")
	b.WriteString(formatNotesSlack(e.Issues))

	return b.String()
}

func formatNotesSlack(s string) string {
	// We wrap to a reasonable width for chat and prefix each line with the note icon.
	// Using a hanging indent keeps wrapped lines readable.
	const wrapWidth = 80
	const indent = "   "

	paragraphs := strings.Split(s, "\n")
	var lines []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(strings.TrimRight(p, "\r"))
		if p == "" {
			continue
		}

		wrapped := text.WrapSoft(p, wrapWidth)
		for i, w := range strings.Split(wrapped, "\n") {
			if i == 0 {
				lines = append(lines, fmt.Sprintf("%s %s", noteIconPrefix, w))
			} else {
				lines = append(lines, indent+w)
			}
		}
	}

	if len(lines) == 0 {
		return fmt.Sprintf("%s %s", noteIconPrefix, none)
	}

	return strings.Join(lines, "\n")
}
