package argsparser

import (
	"errors"
	"strings"
	"testing"
	"time"

	"xn--gckvb8fzb.com/zeit/errs"
)

func TestParseProjectAndTask(t *testing.T) {
	for _, keyword := range projectWords {
		t.Run(keyword, func(t *testing.T) {
			pa, err := Parse("start", []string{keyword, "myproject/mytask"})
			if err != nil {
				t.Fatalf("Parse() = %s", err)
			}

			if pa.ProjectSID != "myproject" {
				t.Errorf("ProjectSID = %q, want %q", pa.ProjectSID, "myproject")
			}
			if pa.TaskSID != "mytask" {
				t.Errorf("TaskSID = %q, want %q", pa.TaskSID, "mytask")
			}
		})
	}
}

func TestParseLowercasesProjectAndTask(t *testing.T) {
	pa, err := Parse("start", []string{"on", "MyProject/MyTask"})
	if err != nil {
		t.Fatalf("Parse() = %s", err)
	}

	if pa.ProjectSID != "myproject" || pa.TaskSID != "mytask" {
		t.Errorf("got %q/%q, want myproject/mytask", pa.ProjectSID, pa.TaskSID)
	}
}

func TestParseRejectsIncompleteProjectAndTask(t *testing.T) {
	tests := [][]string{
		{"on"},
		{"on", "myproject"},
		{"start", "on", "justaproject"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, err := Parse("start", args); errors.Is(
				err, errs.ErrMissingProjectOrTaskSID,
			) == false {
				t.Errorf("Parse(%v) = %v, want ErrMissingProjectOrTaskSID", args, err)
			}
		})
	}
}

func TestParseNote(t *testing.T) {
	for _, keyword := range attributeWords {
		t.Run(keyword, func(t *testing.T) {
			pa, err := Parse("end", []string{keyword, "note", "Issue ID: 123"})
			if err != nil {
				t.Fatalf("Parse() = %s", err)
			}

			if pa.Note != "Issue ID: 123" {
				t.Errorf("Note = %q, want %q", pa.Note, "Issue ID: 123")
			}
		})
	}
}

func TestParseRejectsIncompleteAttributes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{name: "no attribute", args: []string{"with"}, want: errs.ErrMissingAttrOrVal},
		{
			name: "no value",
			args: []string{"with", "note"},
			want: errs.ErrMissingAttrOrVal,
		},
		{
			name: "unknown attribute",
			args: []string{"with", "colour", "red"},
			want: errs.ErrUnknownAttr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Parse("end", test.args); errors.Is(err, test.want) == false {
				t.Errorf("Parse(%v) = %v, want %v", test.args, err, test.want)
			}
		})
	}
}

func TestParseNoiseWordsAreIgnored(t *testing.T) {
	pa, err := Parse("start", []string{
		"block", "working", "work", "wrk", "all", "at", "from",
		"on", "myproject/mytask",
	})
	if err != nil {
		t.Fatalf("Parse() = %s", err)
	}

	if pa.ProjectSID != "myproject" || pa.TaskSID != "mytask" {
		t.Errorf("got %q/%q, want myproject/mytask", pa.ProjectSID, pa.TaskSID)
	}
	if pa.TimestampStart != "" {
		t.Errorf("TimestampStart = %q, want empty", pa.TimestampStart)
	}
}

func TestParseTimestamps(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "start only",
			args:      []string{"5", "minutes", "ago"},
			wantStart: "5 minutes ago",
		},
		{
			name:      "start and end",
			args:      []string{"2", "days", "ago", "until", "now"},
			wantStart: "2 days ago",
			wantEnd:   "now",
		},
		{
			name:      "ended keyword",
			args:      []string{"09:00", "ended", "17:00"},
			wantStart: "09:00",
			wantEnd:   "17:00",
		},
		{
			name:      "from is noise",
			args:      []string{"from", "2", "days", "ago", "until", "now"},
			wantStart: "2 days ago",
			wantEnd:   "now",
		},
		{
			name:    "end only",
			args:    []string{"until", "now"},
			wantEnd: "now",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pa, err := Parse("block", test.args)
			if err != nil {
				t.Fatalf("Parse() = %s", err)
			}

			if pa.TimestampStart != test.wantStart {
				t.Errorf("TimestampStart = %q, want %q",
					pa.TimestampStart, test.wantStart)
			}
			if pa.TimestampEnd != test.wantEnd {
				t.Errorf("TimestampEnd = %q, want %q", pa.TimestampEnd, test.wantEnd)
			}
		})
	}
}

func TestParseTimestampStopsAtKeywords(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStart  string
		wantEnd    string
		wantNote   string
		wantProjct string
		wantTask   string
	}{
		{
			name:       "timestamp then project",
			args:       []string{"5", "minutes", "ago", "on", "myproject/mytask"},
			wantStart:  "5 minutes ago",
			wantProjct: "myproject",
			wantTask:   "mytask",
		},
		{
			name:      "timestamp then note",
			args:      []string{"5", "minutes", "ago", "with", "note", "hello"},
			wantStart: "5 minutes ago",
			wantNote:  "hello",
		},
		{
			name: "project then timestamp then note",
			args: []string{
				"on", "myproject/mytask", "1", "hour", "ago", "with", "note", "hi",
			},
			wantStart:  "1 hour ago",
			wantNote:   "hi",
			wantProjct: "myproject",
			wantTask:   "mytask",
		},
		{
			name: "range then project",
			args: []string{
				"2", "days", "ago", "until", "now", "on", "myproject/mytask",
			},
			wantStart:  "2 days ago",
			wantEnd:    "now",
			wantProjct: "myproject",
			wantTask:   "mytask",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pa, err := Parse("start", test.args)
			if err != nil {
				t.Fatalf("Parse() = %s", err)
			}

			if pa.TimestampStart != test.wantStart {
				t.Errorf("TimestampStart = %q, want %q",
					pa.TimestampStart, test.wantStart)
			}
			if pa.TimestampEnd != test.wantEnd {
				t.Errorf("TimestampEnd = %q, want %q", pa.TimestampEnd, test.wantEnd)
			}
			if pa.Note != test.wantNote {
				t.Errorf("Note = %q, want %q", pa.Note, test.wantNote)
			}
			if pa.ProjectSID != test.wantProjct {
				t.Errorf("ProjectSID = %q, want %q", pa.ProjectSID, test.wantProjct)
			}
			if pa.TaskSID != test.wantTask {
				t.Errorf("TaskSID = %q, want %q", pa.TaskSID, test.wantTask)
			}
		})
	}
}

func TestParseDocumentedExamples(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		want    ParsedArgs
	}{
		{
			name:    "zeit start work with note Hello World on myproject/mytask",
			command: "start",
			args: []string{
				"work", "with", "note", "Hello World", "on", "myproject/mytask",
			},
			want: ParsedArgs{
				ProjectSID: "myproject",
				TaskSID:    "mytask",
				Note:       "Hello World",
			},
		},
		{
			name:    "zeit end with note Issue ID 123 5 minutes ago",
			command: "end",
			args:    []string{"with", "note", "Issue ID 123", "5", "minutes", "ago"},
			want: ParsedArgs{
				Note:         "Issue ID 123",
				TimestampEnd: "5 minutes ago",
			},
		},
		{
			name:    "zeit export all of myproject/mytask from 2 days ago until now",
			command: "export",
			args: []string{
				"all", "of", "myproject/mytask",
				"from", "2", "days", "ago", "until", "now",
			},
			want: ParsedArgs{
				ProjectSID:     "myproject",
				TaskSID:        "mytask",
				TimestampStart: "2 days ago",
				TimestampEnd:   "now",
			},
		},
		{
			name:    "zeit stat on myproject/mytask this week",
			command: "stat",
			args:    []string{"on", "myproject/mytask", "this", "week"},
			want: ParsedArgs{
				ProjectSID:     "myproject",
				TaskSID:        "mytask",
				TimestampStart: "this week",
			},
		},
		{
			name:    "zeit block 01998b32-7f89-7373-a192-56417e0bc89f",
			command: "block",
			args:    []string{"2", "days", "ago"},
			want: ParsedArgs{
				TimestampStart: "2 days ago",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pa, err := Parse(test.command, test.args)
			if err != nil {
				t.Fatalf("Parse() = %s", err)
			}

			if pa.ProjectSID != test.want.ProjectSID ||
				pa.TaskSID != test.want.TaskSID ||
				pa.Note != test.want.Note ||
				pa.TimestampStart != test.want.TimestampStart ||
				pa.TimestampEnd != test.want.TimestampEnd {
				t.Errorf("Parse() = %+v, want %+v", *pa, test.want)
			}
		})
	}
}

func TestParseFullSentences(t *testing.T) {
	note := "Research: New changes in Go 1.27"

	tests := []struct {
		name string
		args []string
		want ParsedArgs
	}{
		{
			name: "project and task only",
			args: []string{"block", "on", "personal/knowledge"},
			want: ParsedArgs{ProjectSID: "personal", TaskSID: "knowledge"},
		},
		{
			name: "with a start",
			args: []string{
				"block", "on", "personal/knowledge", "4", "hours", "ago",
			},
			want: ParsedArgs{
				ProjectSID:     "personal",
				TaskSID:        "knowledge",
				TimestampStart: "4 hours ago",
			},
		},
		{
			name: "with a start and an end",
			args: []string{
				"block", "on", "personal/knowledge",
				"4", "hours", "ago", "ended", "10", "minutes", "ago",
			},
			want: ParsedArgs{
				ProjectSID:     "personal",
				TaskSID:        "knowledge",
				TimestampStart: "4 hours ago",
				TimestampEnd:   "10 minutes ago",
			},
		},
		{
			name: "note before the project",
			args: []string{
				"block", "with", "note", note, "on", "personal/knowledge",
			},
			want: ParsedArgs{
				ProjectSID: "personal",
				TaskSID:    "knowledge",
				Note:       note,
			},
		},
		{
			name: "note before the project and a start",
			args: []string{
				"block", "with", "note", note, "on", "personal/knowledge",
				"4", "hours", "ago",
			},
			want: ParsedArgs{
				ProjectSID:     "personal",
				TaskSID:        "knowledge",
				Note:           note,
				TimestampStart: "4 hours ago",
			},
		},
		{
			name: "note before the project and a range",
			args: []string{
				"block", "with", "note", note, "on", "personal/knowledge",
				"2", "hours", "ago", "ended", "10", "minutes", "ago",
			},
			want: ParsedArgs{
				ProjectSID:     "personal",
				TaskSID:        "knowledge",
				Note:           note,
				TimestampStart: "2 hours ago",
				TimestampEnd:   "10 minutes ago",
			},
		},
		{
			name: "note after a range",
			args: []string{
				"block", "on", "personal/knowledge",
				"2", "hours", "ago", "ended", "10", "minutes", "ago",
				"with", "note", note,
			},
			want: ParsedArgs{
				ProjectSID:     "personal",
				TaskSID:        "knowledge",
				Note:           note,
				TimestampStart: "2 hours ago",
				TimestampEnd:   "10 minutes ago",
			},
		},
		{
			name: "note after a start",
			args: []string{
				"block", "on", "personal/knowledge", "2", "hours", "ago",
				"with", "note", note,
			},
			want: ParsedArgs{
				ProjectSID:     "personal",
				TaskSID:        "knowledge",
				Note:           note,
				TimestampStart: "2 hours ago",
			},
		},
		{
			name: "note after the project",
			args: []string{
				"block", "on", "personal/knowledge", "with", "note", note,
			},
			want: ParsedArgs{
				ProjectSID: "personal",
				TaskSID:    "knowledge",
				Note:       note,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pa, err := Parse("start", test.args)
			if err != nil {
				t.Fatalf("Parse() = %s", err)
			}

			if *pa != test.want {
				t.Errorf("Parse() = %+v, want %+v", *pa, test.want)
			}
		})
	}
}

func TestParseBareTimestampFollowsTheCommand(t *testing.T) {
	tests := []struct {
		command   string
		wantStart string
		wantEnd   string
	}{
		{command: "start", wantStart: "5 minutes ago"},
		{command: "switch", wantStart: "5 minutes ago"},
		{command: "resume", wantStart: "5 minutes ago"},
		{command: "block", wantStart: "5 minutes ago"},
		{command: "stat", wantStart: "5 minutes ago"},
		{command: "export", wantStart: "5 minutes ago"},
		{command: "end", wantEnd: "5 minutes ago"},
		{command: "stop", wantEnd: "5 minutes ago"},
		{command: "pause", wantEnd: "5 minutes ago"},
		{command: "END", wantEnd: "5 minutes ago"},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			pa, err := Parse(test.command, []string{"5", "minutes", "ago"})
			if err != nil {
				t.Fatalf("Parse() = %s", err)
			}

			if pa.TimestampStart != test.wantStart {
				t.Errorf("TimestampStart = %q, want %q",
					pa.TimestampStart, test.wantStart)
			}
			if pa.TimestampEnd != test.wantEnd {
				t.Errorf("TimestampEnd = %q, want %q", pa.TimestampEnd, test.wantEnd)
			}
		})
	}
}

func TestParseEndCommandStillAcceptsAnEndClause(t *testing.T) {
	for _, args := range [][]string{
		{"ended", "5", "minutes", "ago"},
		{"until", "5", "minutes", "ago"},
	} {
		pa, err := Parse("end", args)
		if err != nil {
			t.Fatalf("Parse(%v) = %s", args, err)
		}

		if pa.TimestampEnd != "5 minutes ago" {
			t.Errorf("TimestampEnd = %q, want the end clause", pa.TimestampEnd)
		}
		if pa.TimestampStart != "" {
			t.Errorf("TimestampStart = %q, want empty", pa.TimestampStart)
		}
	}
}

func TestParseRejectsRepeatedClauses(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		want    error
	}{
		{
			name:    "two projects",
			command: "start",
			args:    []string{"on", "alpha/one", "on", "beta/two"},
			want:    errs.ErrRepeatedProjectOrTask,
		},
		{
			name:    "two notes",
			command: "start",
			args:    []string{"with", "note", "first", "with", "note", "second"},
			want:    errs.ErrRepeatedNote,
		},
		{
			name:    "two bare timestamps",
			command: "start",
			args:    []string{"9:00", "on", "alpha/one", "10:00"},
			want:    errs.ErrRepeatedTimestamp,
		},
		{
			name:    "two end clauses",
			command: "start",
			args:    []string{"9:00", "ended", "10:00", "until", "11:00"},
			want:    errs.ErrRepeatedTimestamp,
		},
		{
			name:    "a bare timestamp and an end clause on an end command",
			command: "end",
			args:    []string{"9:00", "ended", "10:00"},
			want:    errs.ErrRepeatedTimestamp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Parse(test.command, test.args); errors.Is(
				err, test.want,
			) == false {
				t.Errorf("Parse(%v) = %v, want %v", test.args, err, test.want)
			}
		})
	}
}

func TestParseRejectsAHalfWrittenProjectAndTask(t *testing.T) {
	for _, args := range [][]string{
		{"on", "alpha/"},
		{"on", "/one"},
		{"on", "/"},
	} {
		if _, err := Parse("start", args); errors.Is(
			err, errs.ErrMissingProjectOrTaskSID,
		) == false {
			t.Errorf("Parse(%v) = %v, want ErrMissingProjectOrTaskSID", args, err)
		}
	}
}

func TestParseAttributeArity(t *testing.T) {
	var valueless int

	attributes["probe"] = attribute{
		Set: func(p *parser, value string) error {
			valueless++
			return nil
		},
	}
	t.Cleanup(func() { delete(attributes, "probe") })

	pa, err := Parse("start", []string{
		"with", "probe", "on", "alpha/one", "with", "probe",
		"with", "note", "kept",
	})
	if err != nil {
		t.Fatalf("Parse() = %s", err)
	}

	if valueless != 2 {
		t.Errorf("the value-less attribute was set %d times, want 2", valueless)
	}
	if pa.ProjectSID != "alpha" || pa.TaskSID != "one" {
		t.Errorf("got %q/%q, want alpha/one", pa.ProjectSID, pa.TaskSID)
	}
	if pa.Note != "kept" {
		t.Errorf("Note = %q, want kept", pa.Note)
	}
}

func TestParseRejectsAnUnknownAttributeBeforeItsValue(t *testing.T) {
	if _, err := Parse("start", []string{"with", "colour"}); errors.Is(
		err, errs.ErrUnknownAttr,
	) == false {
		t.Errorf("Parse() = %v, want ErrUnknownAttr", err)
	}
}

func TestParseEmptyArgs(t *testing.T) {
	pa, err := Parse("block", []string{})
	if err != nil {
		t.Fatalf("Parse() = %s", err)
	}

	if *pa != (ParsedArgs{}) {
		t.Errorf("Parse() = %+v, want a zero value", *pa)
	}
}

func TestProcessParsesTimestamps(t *testing.T) {
	pa := &ParsedArgs{TimestampStart: "2026-08-26 09:00", TimestampEnd: "2026-08-26 17:00"}

	if err := pa.Process(); err != nil {
		t.Fatalf("Process() = %s", err)
	}

	if pa.WasProcessed() == false {
		t.Errorf("WasProcessed() = false, want true")
	}

	start := pa.GetTimestampStart()
	end := pa.GetTimestampEnd()

	if start.Hour() != 9 || end.Hour() != 17 {
		t.Errorf("got %s -> %s, want 09:00 -> 17:00", start, end)
	}
	if start.Before(end) == false {
		t.Errorf("start %s is not before end %s", start, end)
	}
}

func TestProcessExpandsRanges(t *testing.T) {
	pa := &ParsedArgs{TimestampStart: "this week"}

	if err := pa.Process(); err != nil {
		t.Fatalf("Process() = %s", err)
	}

	if pa.GetTimestampStart().IsZero() == true {
		t.Errorf("GetTimestampStart() is zero")
	}
	if pa.GetTimestampEnd().IsZero() == true {
		t.Errorf("a range did not populate the end timestamp")
	}
	if pa.GetTimestampEnd().After(pa.GetTimestampStart()) == false {
		t.Errorf("range end %s is not after start %s",
			pa.GetTimestampEnd(), pa.GetTimestampStart())
	}
}

func TestProcessRejectsUnparseableTimestamps(t *testing.T) {
	for _, pa := range []*ParsedArgs{
		{TimestampStart: "definitely not a date"},
		{TimestampEnd: "definitely not a date"},
	} {
		err := pa.Process()

		parseErr := new(errs.ErrParsingTimestamp)
		if errors.As(err, &parseErr) == false {
			t.Errorf("Process() = %v, want an ErrParsingTimestamp", err)
		}
	}
}

func TestProcessRejectsEndBeforeStart(t *testing.T) {
	pa := &ParsedArgs{
		TimestampStart: "2026-08-26 17:00",
		TimestampEnd:   "2026-08-26 09:00",
	}

	err := pa.Process()

	parseErr := new(errs.ErrParsingTimestamp)
	if errors.As(err, &parseErr) == false {
		t.Fatalf("Process() = %v, want an ErrParsingTimestamp", err)
	}
	if strings.Contains(parseErr.Message, "End is before start") == false {
		t.Errorf("Process() = %q, want a message about the end preceding the start",
			parseErr.Message)
	}
}

func TestProcessValidatesSIDs(t *testing.T) {
	pa := &ParsedArgs{ProjectSID: "my project", TaskSID: "mytask"}

	if err := pa.Process(); errors.Is(err, errs.ErrInvalidSID) == false {
		t.Errorf("Process() = %v, want ErrInvalidSID", err)
	}
}

func TestProcessAcceptsASingleSID(t *testing.T) {
	for _, pa := range []*ParsedArgs{
		{ProjectSID: "myproject"},
		{TaskSID: "mytask"},
	} {
		if err := pa.Process(); err != nil {
			t.Errorf("Process(%+v) = %v, want nil", *pa, err)
		}
	}
}

func TestWasProcessedIsFalseBeforeProcess(t *testing.T) {
	pa := new(ParsedArgs)

	if pa.WasProcessed() == true {
		t.Errorf("WasProcessed() = true on a fresh ParsedArgs")
	}
}

func TestOverrideWith(t *testing.T) {
	pa := &ParsedArgs{
		ProjectSID:     "fromargs",
		TaskSID:        "fromargs",
		Note:           "from args",
		TimestampStart: "from args",
		TimestampEnd:   "from args",
	}

	pa.OverrideWith(&ParsedArgs{ProjectSID: "fromflags", Note: "from flags"})

	if pa.ProjectSID != "fromflags" {
		t.Errorf("ProjectSID = %q, want the flag value", pa.ProjectSID)
	}
	if pa.Note != "from flags" {
		t.Errorf("Note = %q, want the flag value", pa.Note)
	}
	if pa.TaskSID != "fromargs" {
		t.Errorf("TaskSID = %q, want the parsed value", pa.TaskSID)
	}
	if pa.TimestampStart != "from args" || pa.TimestampEnd != "from args" {
		t.Errorf("timestamps were overwritten by empty flags")
	}
}

func TestPOPCombinesParsingAndProcessing(t *testing.T) {
	flags := &ParsedArgs{Note: "from flags"}
	args := []string{"on", "myproject/mytask", "2026-08-26", "09:00"}

	pa, err := POP("start", flags, args, nil)
	if err != nil {
		t.Fatalf("POP() = %s", err)
	}

	if pa.ProjectSID != "myproject" || pa.TaskSID != "mytask" {
		t.Errorf("got %q/%q, want myproject/mytask", pa.ProjectSID, pa.TaskSID)
	}
	if pa.Note != "from flags" {
		t.Errorf("Note = %q, want the flag value", pa.Note)
	}
	if pa.WasProcessed() == false {
		t.Errorf("WasProcessed() = false, want true")
	}
	if pa.GetTimestampStart().IsZero() == true {
		t.Errorf("GetTimestampStart() is zero")
	}
}

func TestPOPReturnsParseErrors(t *testing.T) {
	if _, err := POP("start", new(ParsedArgs), []string{"on"}, nil); errors.Is(
		err, errs.ErrMissingProjectOrTaskSID,
	) == false {
		t.Errorf("POP() = %v, want ErrMissingProjectOrTaskSID", err)
	}
}

func TestGetTimestampsAreZeroBeforeProcess(t *testing.T) {
	pa := &ParsedArgs{TimestampStart: "5 minutes ago"}

	if pa.GetTimestampStart().IsZero() == false {
		t.Errorf("GetTimestampStart() is set before Process()")
	}
	if pa.GetTimestampEnd().IsZero() == false {
		t.Errorf("GetTimestampEnd() is set before Process()")
	}
}

func TestProcessLeavesUnsetTimestampsZero(t *testing.T) {
	pa := &ParsedArgs{ProjectSID: "myproject", TaskSID: "mytask"}

	if err := pa.Process(); err != nil {
		t.Fatalf("Process() = %s", err)
	}

	if pa.GetTimestampStart().IsZero() == false {
		t.Errorf("GetTimestampStart() = %s, want zero", pa.GetTimestampStart())
	}
	if pa.GetTimestampEnd().IsZero() == false {
		t.Errorf("GetTimestampEnd() = %s, want zero", pa.GetTimestampEnd())
	}
}

func TestParsedArgsTimestampsAreIndependentOfWallClock(t *testing.T) {
	pa := &ParsedArgs{TimestampStart: "2026-08-26 09:00"}

	if err := pa.Process(); err != nil {
		t.Fatalf("Process() = %s", err)
	}

	want := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.Local)
	if pa.GetTimestampStart().Equal(want) == false {
		t.Errorf("GetTimestampStart() = %s, want %s", pa.GetTimestampStart(), want)
	}
}
