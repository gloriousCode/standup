package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gloriousCode/reporter/standup"
)

func main() {
	e, err := standup.Prompt(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	date := time.Now().Format("02-01-2006") // DD-MM-YYYY
	outPath := fmt.Sprintf("output-%s.md", date)

	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating output file:", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprintln(f, standup.RenderMarkdown(date, e)); err != nil {
		fmt.Fprintln(os.Stderr, "error writing output file:", err)
		os.Exit(1)
	}

	// Also print a Slack-friendly version to the terminal for easy copy/paste.
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, standup.RenderSlack(date, e))
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "wrote:", outPath)
}
