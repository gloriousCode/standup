package standup

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestPrompt_SingleLineDotTerminators(t *testing.T) {
	in := strings.NewReader("did thing\n.\nno issues\n.\n")
	var out bytes.Buffer

	e, err := Prompt(in, &out)
	if err != nil {
		t.Fatalf("Prompt returned error: %v", err)
	}
	if e.DidToday != "did thing" {
		t.Fatalf("DidToday = %q, want %q", e.DidToday, "did thing")
	}
	if e.Issues != "no issues" {
		t.Fatalf("Issues = %q, want %q", e.Issues, "no issues")
	}

	// Basic sanity that prompts were written.
	gotOut := out.String()
	if !strings.Contains(gotOut, "What did you do today?") {
		t.Fatalf("output missing first prompt, got: %q", gotOut)
	}
	if !strings.Contains(gotOut, "Any issues you need help with?") {
		t.Fatalf("output missing second prompt, got: %q", gotOut)
	}
}

func TestPrompt_EmptyBecomesNone(t *testing.T) {
	in := strings.NewReader("\n.\n\n.\n")
	var out bytes.Buffer

	e, err := Prompt(in, &out)
	if err != nil {
		t.Fatalf("Prompt returned error: %v", err)
	}
	if e.DidToday != none {
		t.Fatalf("DidToday = %q, want %q", e.DidToday, none)
	}
	if e.Issues != none {
		t.Fatalf("Issues = %q, want %q", e.Issues, none)
	}
}

func TestPrompt_EOFMidAnswerAccepted(t *testing.T) {
	in := strings.NewReader("did thing")
	var out bytes.Buffer

	_, err := Prompt(in, &out)
	if err == nil {
		t.Fatalf("expected error because second question can't be answered; got nil")
	}
	if err != io.EOF {
		// bufio.ReadString returns io.EOF in this case.
		t.Fatalf("err = %v, want io.EOF", err)
	}
}

func TestRenderTable_ContainsAnswers(t *testing.T) {
	s := RenderTable(Entry{DidToday: "x", Issues: "y"})
	if !strings.Contains(s, "What I did today") {
		t.Fatalf("render missing label, got:\n%s", s)
	}
	if !strings.Contains(s, "x") || !strings.Contains(s, "y") {
		t.Fatalf("render missing answers, got:\n%s", s)
	}
}

func TestRenderTable_WrapsAndPrefixesNotes(t *testing.T) {
	long := "this is a long note that should wrap across multiple lines so the table doesn't become ridiculously wide"
	s := RenderTable(Entry{DidToday: long, Issues: none})

	// Prefix is applied.
	if !strings.Contains(s, noteIconPrefix+" ") {
		t.Fatalf("render missing note icon prefix %q, got:\n%s", noteIconPrefix, s)
	}

	// Wrapping: expect at least two occurrences of the prefix for the long note.
	if strings.Count(s, noteIconPrefix+" ") < 2 {
		t.Fatalf("expected wrapped note to produce multiple prefixed lines, got:\n%s", s)
	}
}

func TestRenderMarkdown_HiStyleTable(t *testing.T) {
	e := Entry{
		DidToday: "a\nb",
		Issues:   "c",
	}
	md := RenderMarkdown("05-01-2026", e)

	if !strings.Contains(md, "Date: 05-01-2026") {
		t.Fatalf("markdown missing date header, got:\n%s", md)
	}
	if !strings.Contains(md, "| THING") || !strings.Contains(md, "| NOTES") {
		t.Fatalf("markdown missing header row, got:\n%s", md)
	}
	if !strings.Contains(md, "<br/>") {
		t.Fatalf("expected markdown to use <br/> for multi-line notes, got:\n%s", md)
	}
	if strings.Count(md, noteIconPrefix+" ") < 3 {
		t.Fatalf("expected each note line to be prefixed; got:\n%s", md)
	}
}

func TestRenderSlack_BasicFormat(t *testing.T) {
	e := Entry{DidToday: "a\nb", Issues: ""}
	s := RenderSlack("05-01-2026", e)

	if !strings.Contains(s, "Date: 05-01-2026") {
		t.Fatalf("slack render missing date header, got:\n%s", s)
	}
	if !strings.Contains(s, "What I did today") {
		t.Fatalf("slack render missing heading, got:\n%s", s)
	}
	if !strings.Contains(s, "Issues / help needed") {
		t.Fatalf("slack render missing heading, got:\n%s", s)
	}
	if strings.Count(s, noteIconPrefix+" ") < 2 {
		t.Fatalf("slack render missing note prefixes, got:\n%s", s)
	}
}
