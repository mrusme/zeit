package blockFixCmd

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/runtime"
)

var start = time.Date(2026, time.August, 26, 9, 0, 0, 0, time.Local)

func newTestRuntime(t *testing.T) *runtime.Runtime {
	t.Helper()

	t.Setenv(runtime.DATABASE_ENV_VAR, t.TempDir())

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("os.OpenFile() = %s", err)
	}

	previousOut, previousErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = devNull, devNull

	rt := runtime.New(slog.LevelError, out.ColorNever, false)

	os.Stdout, os.Stderr = previousOut, previousErr

	t.Cleanup(func() {
		rt.End()
		devNull.Close()
	})

	return rt
}

func store(t *testing.T, rt *runtime.Runtime, taskSID string, s time.Time, e time.Time) *block.Block {
	t.Helper()

	b, err := block.New("owner")
	if err != nil {
		t.Fatalf("block.New() = %s", err)
	}

	b.ProjectSID = "alpha"
	b.TaskSID = taskSID
	b.TimestampStart = s
	b.TimestampEnd = e

	if err = block.Set(rt.Database, b); err != nil {
		t.Fatalf("block.Set() = %s", err)
	}

	return b
}

func unfinishedIn(t *testing.T, rt *runtime.Runtime) []block.Unfinished {
	t.Helper()

	rows, err := block.List(rt.Database)
	if err != nil {
		t.Fatalf("block.List() = %s", err)
	}

	return block.ListUnfinished(rows, "")
}

func reload(t *testing.T, rt *runtime.Runtime, b *block.Block) *block.Block {
	t.Helper()

	loaded, err := block.Get(rt.Database, b.GetKey())
	if err != nil {
		t.Fatalf("block.Get() = %s", err)
	}

	return loaded
}

func answer(lines string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(lines))
}

func TestFixBlocksAcceptsTheRecommendation(t *testing.T) {
	rt := newTestRuntime(t)

	broken := store(t, rt, "one", start, time.Time{})
	store(t, rt, "two", start.Add(90*time.Minute), start.Add(2*time.Hour))

	if fixed := fixBlocks(rt, answer("\n"), unfinishedIn(t, rt)); fixed != 1 {
		t.Errorf("fixBlocks() = %d, want 1", fixed)
	}

	want := start.Add(90 * time.Minute).Add(-1 * time.Second)
	if got := reload(t, rt, broken).TimestampEnd; got.Equal(want) == false {
		t.Errorf("TimestampEnd = %s, want %s", got, want)
	}
}

func TestFixBlocksAcceptsATypedTimestamp(t *testing.T) {
	rt := newTestRuntime(t)

	broken := store(t, rt, "one", start, time.Time{})
	store(t, rt, "two", start.Add(90*time.Minute), start.Add(2*time.Hour))

	if fixed := fixBlocks(rt, answer("2026-08-26 10:15\n"), unfinishedIn(t, rt)); fixed != 1 {
		t.Errorf("fixBlocks() = %d, want 1", fixed)
	}

	want := time.Date(2026, time.August, 26, 10, 15, 0, 0, time.Local)
	if got := reload(t, rt, broken).TimestampEnd; got.Equal(want) == false {
		t.Errorf("TimestampEnd = %s, want %s", got, want)
	}
}

func TestFixBlocksRetriesAfterAnUnparseableAnswer(t *testing.T) {
	rt := newTestRuntime(t)

	broken := store(t, rt, "one", start, time.Time{})
	store(t, rt, "two", start.Add(90*time.Minute), start.Add(2*time.Hour))

	if fixed := fixBlocks(rt, answer("not a date\n\n"), unfinishedIn(t, rt)); fixed != 1 {
		t.Errorf("fixBlocks() = %d, want 1", fixed)
	}

	if reload(t, rt, broken).TimestampEnd.IsZero() == true {
		t.Errorf("the block was left unfinished after a retry")
	}
}

func TestFixBlocksRetriesAfterAnEndBeforeTheStart(t *testing.T) {
	rt := newTestRuntime(t)

	broken := store(t, rt, "one", start, time.Time{})
	store(t, rt, "two", start.Add(90*time.Minute), start.Add(2*time.Hour))

	if fixed := fixBlocks(rt, answer("2026-08-26 08:00\n\n"), unfinishedIn(t, rt)); fixed != 1 {
		t.Errorf("fixBlocks() = %d, want 1", fixed)
	}

	want := start.Add(90 * time.Minute).Add(-1 * time.Second)
	if got := reload(t, rt, broken).TimestampEnd; got.Equal(want) == false {
		t.Errorf("TimestampEnd = %s, want the recommendation %s", got, want)
	}
}

func TestFixBlocksLeavesABlockWithoutARecommendationAlone(t *testing.T) {
	rt := newTestRuntime(t)

	broken := store(t, rt, "one", start, time.Time{})

	if fixed := fixBlocks(rt, answer("\n"), unfinishedIn(t, rt)); fixed != 0 {
		t.Errorf("fixBlocks() = %d, want 0", fixed)
	}

	if reload(t, rt, broken).TimestampEnd.IsZero() == false {
		t.Errorf("a block without a recommendation was ended anyway")
	}
}

func TestFixBlocksEndsABlockWithoutARecommendationOnRequest(t *testing.T) {
	rt := newTestRuntime(t)

	broken := store(t, rt, "one", start, time.Time{})

	if fixed := fixBlocks(rt, answer("2026-08-26 17:00\n"), unfinishedIn(t, rt)); fixed != 1 {
		t.Errorf("fixBlocks() = %d, want 1", fixed)
	}

	want := time.Date(2026, time.August, 26, 17, 0, 0, 0, time.Local)
	if got := reload(t, rt, broken).TimestampEnd; got.Equal(want) == false {
		t.Errorf("TimestampEnd = %s, want %s", got, want)
	}
}

func TestFixBlocksStopsAtEndOfInput(t *testing.T) {
	rt := newTestRuntime(t)

	first := store(t, rt, "one", start, time.Time{})
	second := store(t, rt, "two", start.Add(2*time.Hour), time.Time{})
	store(t, rt, "three", start.Add(4*time.Hour), start.Add(5*time.Hour))

	if fixed := fixBlocks(rt, answer(""), unfinishedIn(t, rt)); fixed != 0 {
		t.Errorf("fixBlocks() = %d, want 0", fixed)
	}

	if reload(t, rt, first).TimestampEnd.IsZero() == false ||
		reload(t, rt, second).TimestampEnd.IsZero() == false {
		t.Errorf("a block was ended despite there being no input")
	}
}

func TestFixBlocksHandlesSeveralBlocks(t *testing.T) {
	rt := newTestRuntime(t)

	first := store(t, rt, "one", start, time.Time{})
	second := store(t, rt, "two", start.Add(2*time.Hour), time.Time{})
	store(t, rt, "three", start.Add(4*time.Hour), start.Add(5*time.Hour))

	if fixed := fixBlocks(rt, answer("\n2026-08-26 12:30\n"), unfinishedIn(t, rt)); fixed != 2 {
		t.Errorf("fixBlocks() = %d, want 2", fixed)
	}

	firstWant := start.Add(2 * time.Hour).Add(-1 * time.Second)
	if got := reload(t, rt, first).TimestampEnd; got.Equal(firstWant) == false {
		t.Errorf("first TimestampEnd = %s, want %s", got, firstWant)
	}

	secondWant := time.Date(2026, time.August, 26, 12, 30, 0, 0, time.Local)
	if got := reload(t, rt, second).TimestampEnd; got.Equal(secondWant) == false {
		t.Errorf("second TimestampEnd = %s, want %s", got, secondWant)
	}
}

func TestFixBlocksOnAnEmptyList(t *testing.T) {
	rt := newTestRuntime(t)

	if fixed := fixBlocks(rt, answer(""), nil); fixed != 0 {
		t.Errorf("fixBlocks() = %d, want 0", fixed)
	}
}

func TestAskForEndAcceptsTheRecommendation(t *testing.T) {
	rt := newTestRuntime(t)

	recommended := start.Add(time.Hour)

	got, err := askForEnd(rt, answer("\n"), recommended)
	if err != nil {
		t.Fatalf("askForEnd() = %s", err)
	}

	if got.Equal(recommended) == false {
		t.Errorf("askForEnd() = %s, want %s", got, recommended)
	}
}

func TestAskForEndTrimsTheAnswer(t *testing.T) {
	rt := newTestRuntime(t)

	got, err := askForEnd(rt, answer("   \n"), start.Add(time.Hour))
	if err != nil {
		t.Fatalf("askForEnd() = %s", err)
	}

	if got.Equal(start.Add(time.Hour)) == false {
		t.Errorf("askForEnd() = %s, want the recommendation", got)
	}
}

func TestAskForEndReportsEndOfInput(t *testing.T) {
	rt := newTestRuntime(t)

	if _, err := askForEnd(rt, answer(""), start.Add(time.Hour)); err == nil {
		t.Errorf("askForEnd() = nil error at end of input")
	}
}

func TestAskForEndAcceptsAnAnswerWithoutATrailingNewline(t *testing.T) {
	rt := newTestRuntime(t)

	got, err := askForEnd(rt, answer("2026-08-26 17:00"), time.Time{})
	if err != nil {
		t.Fatalf("askForEnd() = %s", err)
	}

	want := time.Date(2026, time.August, 26, 17, 0, 0, 0, time.Local)
	if got.Equal(want) == false {
		t.Errorf("askForEnd() = %s, want %s", got, want)
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{count: 0, want: "blocks"},
		{count: 1, want: "block"},
		{count: 2, want: "blocks"},
	}

	for _, test := range tests {
		if got := pluralize(test.count, "block", "blocks"); got != test.want {
			t.Errorf("pluralize(%d) = %q, want %q", test.count, got, test.want)
		}
	}
}
