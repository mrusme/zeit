package block

import (
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"xn--gckvb8fzb.com/zeit/database"
	"xn--gckvb8fzb.com/zeit/errs"
	"xn--gckvb8fzb.com/zeit/helpers/argsparser"
	"xn--gckvb8fzb.com/zeit/helpers/log"
	"xn--gckvb8fzb.com/zeit/models/activeblock"
)

func newTestDatabase(t *testing.T) *database.Database {
	t.Helper()

	db, err := database.New(log.New(slog.LevelError), "", false)
	if err != nil {
		t.Fatalf("database.New() = %s", err)
	}

	t.Cleanup(db.Close)

	return db
}

func newBlock(t *testing.T, projectSID, taskSID string) *Block {
	t.Helper()

	b, err := New("owner")
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	b.ProjectSID = projectSID
	b.TaskSID = taskSID

	return b
}

func mustStart(t *testing.T, db *database.Database, projectSID, taskSID string) *Block {
	t.Helper()

	b, err := Start(db, newBlock(t, projectSID, taskSID))
	if err != nil {
		t.Fatalf("Start(%s/%s) = %s", projectSID, taskSID, err)
	}

	return b
}

func TestNew(t *testing.T) {
	b, err := New("owner")
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if b.OwnerKey != "owner" {
		t.Errorf("OwnerKey = %q, want owner", b.OwnerKey)
	}
	if strings.HasPrefix(b.GetKey(), "block:") == false {
		t.Errorf("GetKey() = %q, want a block: prefix", b.GetKey())
	}
}

func TestStartMarksTheBlockActive(t *testing.T) {
	db := newTestDatabase(t)

	started := mustStart(t, db, "alpha", "one")

	if started.TimestampStart.IsZero() == true {
		t.Errorf("Start() left TimestampStart zero")
	}
	if started.TimestampEnd.IsZero() == false {
		t.Errorf("Start() set TimestampEnd on a running block")
	}

	found, _, running, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}

	if found == false {
		t.Fatalf("GetActive() found no running block")
	}
	if running.GetKey() != started.GetKey() {
		t.Errorf("GetActive() = %q, want %q", running.GetKey(), started.GetKey())
	}
}

func TestStartKeepsAUserSuppliedStart(t *testing.T) {
	db := newTestDatabase(t)

	want := time.Now().Add(-90 * time.Minute).Truncate(time.Second)

	b := newBlock(t, "alpha", "one")
	b.TimestampStart = want

	started, err := Start(db, b)
	if err != nil {
		t.Fatalf("Start() = %s", err)
	}

	if started.TimestampStart.Equal(want) == false {
		t.Errorf("TimestampStart = %s, want %s", started.TimestampStart, want)
	}
}

func TestStartWithAnEndStoresAFinishedBlock(t *testing.T) {
	db := newTestDatabase(t)

	start := time.Now().Add(-2 * time.Hour)
	end := time.Now().Add(-time.Hour)

	b := newBlock(t, "alpha", "one")
	b.TimestampStart = start
	b.TimestampEnd = end

	if _, err := Start(db, b); err != nil {
		t.Fatalf("Start() = %s", err)
	}

	found, _, _, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}

	if found == true {
		t.Errorf("a block created with an end timestamp was marked as running")
	}
}

func TestStartWithAnEndWhileTrackingIsRejected(t *testing.T) {
	db := newTestDatabase(t)

	running := mustStart(t, db, "alpha", "one")

	b := newBlock(t, "beta", "two")
	b.TimestampStart = time.Now().Add(-3 * time.Hour)
	b.TimestampEnd = time.Now().Add(-2 * time.Hour)

	if _, err := Start(db, b); errors.Is(err, errs.ErrInvalidTimestampEnd) == false {
		t.Errorf("Start() = %v, want ErrInvalidTimestampEnd", err)
	}

	found, _, active, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}
	if found == false || active.GetKey() != running.GetKey() {
		t.Errorf("the rejected insert disturbed the running block")
	}

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}
	if len(rows) != 1 {
		t.Errorf("List() returned %d blocks, want only the running one", len(rows))
	}
}

func TestStartWithAnEndOnAnIdleDatabase(t *testing.T) {
	db := newTestDatabase(t)

	b := newBlock(t, "beta", "two")
	b.TimestampStart = time.Now().Add(-3 * time.Hour)
	b.TimestampEnd = time.Now().Add(-2 * time.Hour)

	if _, err := Start(db, b); err != nil {
		t.Fatalf("Start() = %s", err)
	}

	found, _, _, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}
	if found == true {
		t.Errorf("inserting a past entry started tracking")
	}
}

func TestResumeAfterASwitchResumesTheLastTask(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")
	mustStart(t, db, "beta", "two")

	if _, err := End(db, newBlock(t, "", "")); err != nil {
		t.Fatalf("End() = %s", err)
	}

	resumed, err := Resume(db, newBlock(t, "", ""))
	if err != nil {
		t.Fatalf("Resume() = %s", err)
	}

	if resumed.ProjectSID != "beta" || resumed.TaskSID != "two" {
		t.Errorf("Resume() = %s/%s, want beta/two",
			resumed.ProjectSID, resumed.TaskSID)
	}
}

func TestStartRejectsAnEndBeforeTheStart(t *testing.T) {
	db := newTestDatabase(t)

	b := newBlock(t, "alpha", "one")
	b.TimestampStart = time.Now()
	b.TimestampEnd = time.Now().Add(-time.Hour)

	if _, err := Start(db, b); errors.Is(err, errs.ErrInvalidTimestampEnd) == false {
		t.Errorf("Start() = %v, want ErrInvalidTimestampEnd", err)
	}
}

func TestStartRejectsInvalidSIDs(t *testing.T) {
	db := newTestDatabase(t)

	tests := []struct {
		name       string
		projectSID string
		taskSID    string
	}{
		{name: "missing project", projectSID: "", taskSID: "one"},
		{name: "missing task", projectSID: "alpha", taskSID: ""},
		{name: "invalid project", projectSID: "al pha", taskSID: "one"},
		{name: "invalid task", projectSID: "alpha", taskSID: "o ne"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Start(
				db, newBlock(t, test.projectSID, test.taskSID),
			); err == nil {
				t.Errorf("Start() = nil, want an error")
			}
		})
	}
}

func TestStartRejectsTheSameProjectAndTaskTwice(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")

	if _, err := Start(db, newBlock(t, "alpha", "one")); errors.Is(
		err, errs.ErrEndStartSIDsIdentical,
	) == false {
		t.Errorf("Start() = %v, want ErrEndStartSIDsIdentical", err)
	}
}

func TestStartClosesTheRunningBlockContiguously(t *testing.T) {
	db := newTestDatabase(t)

	first := mustStart(t, db, "alpha", "one")
	second := mustStart(t, db, "beta", "two")

	closed, err := Get(db, first.GetKey())
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if closed.TimestampEnd.IsZero() == true {
		t.Fatalf("the previous block was left running")
	}
	if closed.TimestampEnd.Equal(second.TimestampStart) == false {
		t.Errorf("previous block ends at %s, the new one starts at %s",
			closed.TimestampEnd, second.TimestampStart)
	}
}

func TestStartInQuickSuccession(t *testing.T) {
	db := newTestDatabase(t)

	for _, sid := range []string{"one", "two", "three", "four"} {
		if _, err := Start(db, newBlock(t, "alpha", sid)); err != nil {
			t.Fatalf("Start(alpha/%s) = %s", sid, err)
		}
	}

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	if len(rows) != 4 {
		t.Errorf("List() returned %d blocks, want 4", len(rows))
	}
}

func TestSwitchBehavesLikeStart(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")

	switched, err := Switch(db, newBlock(t, "beta", "two"))
	if err != nil {
		t.Fatalf("Switch() = %s", err)
	}

	found, _, running, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}

	if found == false || running.GetKey() != switched.GetKey() {
		t.Errorf("Switch() did not leave the new block running")
	}
}

func TestEnd(t *testing.T) {
	db := newTestDatabase(t)

	started := mustStart(t, db, "alpha", "one")

	ended, err := End(db, newBlock(t, "", ""))
	if err != nil {
		t.Fatalf("End() = %s", err)
	}

	if ended.GetKey() != started.GetKey() {
		t.Errorf("End() = %q, want %q", ended.GetKey(), started.GetKey())
	}
	if ended.TimestampEnd.IsZero() == true {
		t.Errorf("End() left TimestampEnd zero")
	}

	found, _, _, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}
	if found == true {
		t.Errorf("GetActive() still reports a running block")
	}
}

func TestEndOnTheRunningProjectAndTask(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")

	if _, err := End(db, newBlock(t, "alpha", "one")); err != nil {
		t.Errorf("End() = %s, want the running block to be ended", err)
	}
}

func TestEndWithNothingRunning(t *testing.T) {
	db := newTestDatabase(t)

	if _, err := End(db, newBlock(t, "", "")); errors.Is(
		err, errs.ErrNothingToEnd,
	) == false {
		t.Errorf("End() = %v, want ErrNothingToEnd", err)
	}
}

func TestEndAppliesANote(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")

	b := newBlock(t, "", "")
	b.Note = "wrapped up"

	ended, err := End(db, b)
	if err != nil {
		t.Fatalf("End() = %s", err)
	}

	if ended.Note != "wrapped up" {
		t.Errorf("Note = %q, want the supplied note", ended.Note)
	}

	loaded, err := Get(db, ended.GetKey())
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}
	if loaded.Note != "wrapped up" {
		t.Errorf("the note was not persisted")
	}
}

func TestEndRejectsAnEndBeforeTheStart(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")

	b := newBlock(t, "", "")
	b.TimestampEnd = time.Now().Add(-time.Hour)

	if _, err := End(db, b); errors.Is(err, errs.ErrInvalidTimestampEnd) == false {
		t.Errorf("End() = %v, want ErrInvalidTimestampEnd", err)
	}
}

func TestResume(t *testing.T) {
	db := newTestDatabase(t)

	first := mustStart(t, db, "alpha", "one")

	if _, err := End(db, newBlock(t, "", "")); err != nil {
		t.Fatalf("End() = %s", err)
	}

	resumed, err := Resume(db, newBlock(t, "", ""))
	if err != nil {
		t.Fatalf("Resume() = %s", err)
	}

	if resumed.ProjectSID != first.ProjectSID || resumed.TaskSID != first.TaskSID {
		t.Errorf("Resume() = %s/%s, want %s/%s",
			resumed.ProjectSID, resumed.TaskSID, first.ProjectSID, first.TaskSID)
	}
	if resumed.GetKey() == first.GetKey() {
		t.Errorf("Resume() reused the previous block instead of starting a new one")
	}
}

func TestResumeWhileRunning(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")

	if _, err := Resume(db, newBlock(t, "", "")); errors.Is(
		err, errs.ErrAlreadyRunning,
	) == false {
		t.Errorf("Resume() = %v, want ErrAlreadyRunning", err)
	}
}

func TestResumeWithNothingToResume(t *testing.T) {
	db := newTestDatabase(t)

	if _, err := Resume(db, newBlock(t, "", "")); errors.Is(
		err, errs.ErrNothingToResume,
	) == false {
		t.Errorf("Resume() = %v, want ErrNothingToResume", err)
	}
}

func TestGetActiveOnAnEmptyDatabase(t *testing.T) {
	db := newTestDatabase(t)

	found, ab, running, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}

	if found == true {
		t.Errorf("GetActive() = true on an empty database")
	}
	if ab == nil {
		t.Errorf("GetActive() returned a nil ActiveBlock")
	}
	if running != nil {
		t.Errorf("GetActive() returned a block while reporting none")
	}
}

func TestGetActiveClearsADanglingReference(t *testing.T) {
	db := newTestDatabase(t)

	ab, err := activeblock.Get(db)
	if err != nil {
		t.Fatalf("activeblock.Get() = %s", err)
	}
	ab.SetActiveBlockKey("block:does-not-exist")
	if err = activeblock.Set(db, ab); err != nil {
		t.Fatalf("activeblock.Set() = %s", err)
	}

	found, _, _, err := GetActive(db)
	if err != nil {
		t.Fatalf("GetActive() = %s", err)
	}
	if found == true {
		t.Errorf("GetActive() = true for a block that no longer exists")
	}

	reloaded, err := activeblock.Get(db)
	if err != nil {
		t.Fatalf("activeblock.Get() = %s", err)
	}
	if reloaded.HasActiveBlockKey() == true {
		t.Errorf("the dangling reference was not cleared")
	}
}

func TestGetMissingBlock(t *testing.T) {
	db := newTestDatabase(t)

	if _, err := Get(db, "block:missing"); errors.Is(
		err, errs.ErrKeyNotFound,
	) == false {
		t.Errorf("Get() = %v, want ErrKeyNotFound", err)
	}
}

func TestListForProjectSIDAndTaskSID(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")
	mustStart(t, db, "alpha", "two")
	mustStart(t, db, "beta", "one")

	rows, err := ListForProjectSID(db, "alpha")
	if err != nil {
		t.Fatalf("ListForProjectSID() = %s", err)
	}
	if len(rows) != 2 {
		t.Errorf("ListForProjectSID() returned %d blocks, want 2", len(rows))
	}

	rows, err = ListForProjectTaskSID(db, "alpha", "one")
	if err != nil {
		t.Fatalf("ListForProjectTaskSID() = %s", err)
	}
	if len(rows) != 1 {
		t.Errorf("ListForProjectTaskSID() returned %d blocks, want 1", len(rows))
	}
}

func TestGroupByProjectTaskSID(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")
	mustStart(t, db, "alpha", "two")
	mustStart(t, db, "alpha", "one")
	mustStart(t, db, "beta", "one")

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	grouped := GroupByProjectTaskSID(rows)

	if len(grouped["alpha"]["one"]) != 2 {
		t.Errorf("alpha/one holds %d blocks, want 2", len(grouped["alpha"]["one"]))
	}
	if len(grouped["alpha"]["two"]) != 1 {
		t.Errorf("alpha/two holds %d blocks, want 1", len(grouped["alpha"]["two"]))
	}
	if len(grouped["beta"]["one"]) != 1 {
		t.Errorf("beta/one holds %d blocks, want 1", len(grouped["beta"]["one"]))
	}
	if len(grouped["nothere"]["one"]) != 0 {
		t.Errorf("an absent project returned blocks")
	}
	if len(grouped["alpha"]["nothere"]) != 0 {
		t.Errorf("an absent task returned blocks")
	}

	for projectSID, tasks := range grouped {
		for taskSID, blocks := range tasks {
			for key, b := range blocks {
				if b.ProjectSID != projectSID || b.TaskSID != taskSID {
					t.Errorf("block %q is grouped under %s/%s but belongs to %s/%s",
						key, projectSID, taskSID, b.ProjectSID, b.TaskSID)
				}
				if b.GetKey() != key {
					t.Errorf("block %q carries key %q", key, b.GetKey())
				}
			}
		}
	}
}

func TestGroupByProjectTaskSIDMatchesListForProjectTaskSID(t *testing.T) {
	db := newTestDatabase(t)

	mustStart(t, db, "alpha", "one")
	mustStart(t, db, "alpha", "two")
	mustStart(t, db, "beta", "one")

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	grouped := GroupByProjectTaskSID(rows)

	pairs := [][2]string{
		{"alpha", "one"},
		{"alpha", "two"},
		{"beta", "one"},
		{"beta", "two"},
	}

	for _, pair := range pairs {
		listed, err := ListForProjectTaskSID(db, pair[0], pair[1])
		if err != nil {
			t.Fatalf("ListForProjectTaskSID() = %s", err)
		}

		if len(listed) != len(grouped[pair[0]][pair[1]]) {
			t.Errorf("%s/%s: listing has %d blocks, grouping has %d",
				pair[0], pair[1], len(listed), len(grouped[pair[0]][pair[1]]))
		}
		for key := range listed {
			if grouped[pair[0]][pair[1]][key] == nil {
				t.Errorf("%s/%s: block %q is missing from the grouping",
					pair[0], pair[1], key)
			}
		}
	}
}

func TestGroupByProjectTaskSIDOnAnEmptyMap(t *testing.T) {
	if len(GroupByProjectTaskSID(map[string]*Block{})) != 0 {
		t.Errorf("GroupByProjectTaskSID() = a non-empty grouping")
	}
}

func unfinishedBlock(t *testing.T, projectSID, taskSID string, start time.Time) *Block {
	t.Helper()

	b := newBlock(t, projectSID, taskSID)
	b.TimestampStart = start

	return b
}

func finishedBlock(
	t *testing.T,
	projectSID string,
	taskSID string,
	start time.Time,
	end time.Time,
) *Block {
	t.Helper()

	b := unfinishedBlock(t, projectSID, taskSID, start)
	b.TimestampEnd = end

	return b
}

func asRows(bs ...*Block) map[string]*Block {
	rows := make(map[string]*Block)

	for _, b := range bs {
		rows[b.GetKey()] = b
	}

	return rows
}

func TestListUnfinishedOnACleanDatabase(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	rows := asRows(
		finishedBlock(t, "alpha", "one", start, start.Add(time.Hour)),
		finishedBlock(t, "alpha", "two", start.Add(time.Hour), start.Add(2*time.Hour)),
	)

	if got := ListUnfinished(rows, ""); len(got) != 0 {
		t.Errorf("ListUnfinished() = %d entries, want none", len(got))
	}
}

func TestListUnfinishedSkipsTheRunningBlock(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	running := unfinishedBlock(t, "alpha", "one", start)
	rows := asRows(running)

	if got := ListUnfinished(rows, running.GetKey()); len(got) != 0 {
		t.Errorf("ListUnfinished() = %d entries, want the running block to be skipped",
			len(got))
	}

	if got := ListUnfinished(rows, ""); len(got) != 1 {
		t.Errorf("ListUnfinished() = %d entries, want the same block once it is "+
			"no longer running", len(got))
	}
}

func TestListUnfinishedRecommendsTheNextStartMinusASecond(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)
	next := start.Add(90 * time.Minute)

	rows := asRows(
		unfinishedBlock(t, "alpha", "one", start),
		finishedBlock(t, "alpha", "two", next, next.Add(time.Hour)),
	)

	got := ListUnfinished(rows, "")
	if len(got) != 1 {
		t.Fatalf("ListUnfinished() = %d entries, want 1", len(got))
	}

	want := next.Add(-1 * time.Second)
	if got[0].RecommendedEnd.Equal(want) == false {
		t.Errorf("RecommendedEnd = %s, want %s", got[0].RecommendedEnd, want)
	}
}

func TestListUnfinishedWithoutAFollowingBlock(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	rows := asRows(
		finishedBlock(t, "alpha", "one", start.Add(-2*time.Hour), start.Add(-time.Hour)),
		unfinishedBlock(t, "alpha", "two", start),
	)

	got := ListUnfinished(rows, "")
	if len(got) != 1 {
		t.Fatalf("ListUnfinished() = %d entries, want 1", len(got))
	}

	if got[0].RecommendedEnd.IsZero() == false {
		t.Errorf("RecommendedEnd = %s, want no recommendation",
			got[0].RecommendedEnd)
	}
}

func TestListUnfinishedWithoutARecommendationThatWouldPrecedeTheStart(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	rows := asRows(
		unfinishedBlock(t, "alpha", "one", start),
		finishedBlock(t, "alpha", "two",
			start.Add(500*time.Millisecond), start.Add(time.Hour)),
	)

	got := ListUnfinished(rows, "")
	if len(got) != 1 {
		t.Fatalf("ListUnfinished() = %d entries, want 1", len(got))
	}

	if got[0].RecommendedEnd.IsZero() == false {
		t.Errorf("RecommendedEnd = %s, want no recommendation when it would "+
			"precede the start", got[0].RecommendedEnd)
	}
}

func TestListUnfinishedIsOrderedByStart(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	rows := asRows(
		unfinishedBlock(t, "alpha", "third", start.Add(4*time.Hour)),
		unfinishedBlock(t, "alpha", "first", start),
		unfinishedBlock(t, "alpha", "second", start.Add(2*time.Hour)),
		finishedBlock(t, "alpha", "last",
			start.Add(6*time.Hour), start.Add(7*time.Hour)),
	)

	got := ListUnfinished(rows, "")
	if len(got) != 3 {
		t.Fatalf("ListUnfinished() = %d entries, want 3", len(got))
	}

	want := []string{"first", "second", "third"}
	for i, taskSID := range want {
		if got[i].Block.TaskSID != taskSID {
			t.Errorf("entry %d is %q, want %q", i, got[i].Block.TaskSID, taskSID)
		}
	}
}

func TestListUnfinishedRecommendsAcrossConsecutiveGaps(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	rows := asRows(
		unfinishedBlock(t, "alpha", "one", start),
		unfinishedBlock(t, "alpha", "two", start.Add(2*time.Hour)),
		finishedBlock(t, "alpha", "three",
			start.Add(5*time.Hour), start.Add(6*time.Hour)),
	)

	got := ListUnfinished(rows, "")
	if len(got) != 2 {
		t.Fatalf("ListUnfinished() = %d entries, want 2", len(got))
	}

	first := start.Add(2 * time.Hour).Add(-1 * time.Second)
	if got[0].RecommendedEnd.Equal(first) == false {
		t.Errorf("first RecommendedEnd = %s, want %s", got[0].RecommendedEnd, first)
	}

	second := start.Add(5 * time.Hour).Add(-1 * time.Second)
	if got[1].RecommendedEnd.Equal(second) == false {
		t.Errorf("second RecommendedEnd = %s, want %s", got[1].RecommendedEnd, second)
	}
}

func TestListUnfinishedRecommendationsAreValidEnds(t *testing.T) {
	db := newTestDatabase(t)

	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	rows := asRows(
		unfinishedBlock(t, "alpha", "one", start),
		finishedBlock(t, "alpha", "two",
			start.Add(2*time.Hour), start.Add(3*time.Hour)),
	)

	for _, u := range ListUnfinished(rows, "") {
		u.Block.TimestampEnd = u.RecommendedEnd

		if err := Set(db, u.Block); err != nil {
			t.Errorf("the recommendation %s is not a valid end: %s",
				u.RecommendedEnd, err)
		}
	}
}

func TestListUnfinishedOnAnEmptyMap(t *testing.T) {
	if got := ListUnfinished(map[string]*Block{}, ""); len(got) != 0 {
		t.Errorf("ListUnfinished() = %d entries, want none", len(got))
	}
}

func TestGetNotePreview(t *testing.T) {
	tests := []struct {
		name   string
		note   string
		length int
		want   string
	}{
		{name: "empty", note: "", want: "// no note added"},
		{name: "short", note: "a note", want: "a note"},
		{
			name: "newlines are replaced",
			note: "first\nsecond",
			want: "first⏎second",
		},
		{
			name:   "truncated",
			note:   strings.Repeat("a", 20),
			length: 10,
			want:   strings.Repeat("a", 10) + "...",
		},
		{
			name:   "exactly the limit",
			note:   strings.Repeat("a", 10),
			length: 10,
			want:   strings.Repeat("a", 10),
		},
		{
			name:   "multibyte is cut by rune",
			note:   strings.Repeat("日", 20),
			length: 10,
			want:   strings.Repeat("日", 10) + "...",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := GetNotePreview(test.note, test.length)
			if got != test.want {
				t.Errorf("GetNotePreview() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestGetNotePreviewKeepsValidUTF8(t *testing.T) {
	for prefix := range 8 {
		note := strings.Repeat("x", prefix) + strings.Repeat("日", 80)

		if utf8.ValidString(GetNotePreview(note, 0)) == false {
			t.Errorf("GetNotePreview() produced invalid UTF-8 for a %d character prefix",
				prefix)
		}
	}
}

func TestGetNotePreviewDefaultLength(t *testing.T) {
	note := strings.Repeat("a", 200)

	got := GetNotePreview(note, 0)

	if got != strings.Repeat("a", 73)+"..." {
		t.Errorf("GetNotePreview() = %q, want a 73 character preview", got)
	}
}

func TestFromProcessedArgs(t *testing.T) {
	pargs, err := argsparser.POP("start", new(argsparser.ParsedArgs), []string{
		"on", "alpha/one", "with", "note", "hello",
	}, nil)
	if err != nil {
		t.Fatalf("POP() = %s", err)
	}

	b := newBlock(t, "", "")
	if err = b.FromProcessedArgs(pargs); err != nil {
		t.Fatalf("FromProcessedArgs() = %s", err)
	}

	if b.ProjectSID != "alpha" || b.TaskSID != "one" {
		t.Errorf("got %s/%s, want alpha/one", b.ProjectSID, b.TaskSID)
	}
	if b.Note != "hello" {
		t.Errorf("Note = %q, want hello", b.Note)
	}
}

func TestFromProcessedArgsRejectsUnprocessedArgs(t *testing.T) {
	b := newBlock(t, "", "")

	if err := b.FromProcessedArgs(new(argsparser.ParsedArgs)); err == nil {
		t.Errorf("FromProcessedArgs() = nil, want an error")
	}
}

func TestFromProcessedArgsLeavesUnsetFieldsAlone(t *testing.T) {
	pargs, err := argsparser.POP("start", new(argsparser.ParsedArgs), []string{}, nil)
	if err != nil {
		t.Fatalf("POP() = %s", err)
	}

	b := newBlock(t, "alpha", "one")
	b.Note = "keep me"

	if err = b.FromProcessedArgs(pargs); err != nil {
		t.Fatalf("FromProcessedArgs() = %s", err)
	}

	if b.ProjectSID != "alpha" || b.TaskSID != "one" || b.Note != "keep me" {
		t.Errorf("FromProcessedArgs() overwrote existing values: %+v", b)
	}
}

func TestSetRejectsAnInvalidBlock(t *testing.T) {
	db := newTestDatabase(t)

	b := newBlock(t, "alpha", "one")
	b.TimestampStart = time.Now()
	b.Note = strings.Repeat("a", 65537)

	if err := Set(db, b); errors.Is(err, errs.ErrNoteTooLarge) == false {
		t.Errorf("Set() = %v, want ErrNoteTooLarge", err)
	}
}

func TestSetAndGetRoundTrip(t *testing.T) {
	db := newTestDatabase(t)

	b := newBlock(t, "alpha", "one")
	b.TimestampStart = time.Now().Add(-time.Hour).Truncate(time.Second)
	b.TimestampEnd = time.Now().Truncate(time.Second)
	b.Note = "a note"

	if err := Set(db, b); err != nil {
		t.Fatalf("Set() = %s", err)
	}

	loaded, err := Get(db, b.GetKey())
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if loaded.ProjectSID != "alpha" || loaded.TaskSID != "one" {
		t.Errorf("got %s/%s, want alpha/one", loaded.ProjectSID, loaded.TaskSID)
	}
	if loaded.Note != "a note" {
		t.Errorf("Note = %q, want a note", loaded.Note)
	}
	if loaded.TimestampStart.Equal(b.TimestampStart) == false {
		t.Errorf("TimestampStart = %s, want %s", loaded.TimestampStart, b.TimestampStart)
	}
	if loaded.TimestampEnd.Equal(b.TimestampEnd) == false {
		t.Errorf("TimestampEnd = %s, want %s", loaded.TimestampEnd, b.TimestampEnd)
	}
}
