package statCmd

import (
	"errors"
	"testing"
	"time"

	"xn--gckvb8fzb.com/zeit/errs"
	"xn--gckvb8fzb.com/zeit/helpers/timestamp"
	"xn--gckvb8fzb.com/zeit/models/block"
)

func periodKey(t *testing.T, name string, at time.Time) string {
	t.Helper()

	period, err := timestamp.GetPeriod(name)
	if err != nil {
		t.Fatalf("timestamp.GetPeriod(%q) = %s", name, err)
	}

	return period.Key(at)
}

func newBlock(t *testing.T, projectSID, taskSID string, start, end time.Time) *block.Block {
	t.Helper()

	b, err := block.New("owner")
	if err != nil {
		t.Fatalf("block.New() = %s", err)
	}

	b.ProjectSID = projectSID
	b.TaskSID = taskSID
	b.TimestampStart = start
	b.TimestampEnd = end

	return b
}

func TestAggregateDurationsByDay(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	bs := []*block.Block{
		newBlock(t, "alpha", "one", start, start.Add(time.Hour)),
		newBlock(t, "alpha", "one", start.Add(2*time.Hour), start.Add(3*time.Hour)),
		newBlock(t, "alpha", "two", start, start.Add(30*time.Minute)),
		newBlock(t, "beta", "one", start.AddDate(0, 0, 1), start.AddDate(0, 0, 1).Add(time.Hour)),
	}

	stats, err := aggregateDurations(bs, "day", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if got := stats["alpha"]["one"]["2026-08-26"]; got != 2*time.Hour {
		t.Errorf("alpha/one on 2026-08-26 = %s, want 2h", got)
	}
	if got := stats["alpha"]["two"]["2026-08-26"]; got != 30*time.Minute {
		t.Errorf("alpha/two on 2026-08-26 = %s, want 30m", got)
	}
	if got := stats["beta"]["one"]["2026-08-27"]; got != time.Hour {
		t.Errorf("beta/one on 2026-08-27 = %s, want 1h", got)
	}
	if got := stats["*"]["*"]["*"]; got != 3*time.Hour+30*time.Minute {
		t.Errorf("total = %s, want 3h30m", got)
	}
}

func TestAggregateDurationsByWeekAndMonth(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)

	bs := []*block.Block{
		newBlock(t, "alpha", "one", start, start.Add(time.Hour)),
		newBlock(t, "alpha", "one", start.AddDate(0, 0, 1), start.AddDate(0, 0, 1).Add(time.Hour)),
	}

	weekly, err := aggregateDurations(bs, "week", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}
	if got := weekly["alpha"]["one"][periodKey(t, "week", start)]; got != 2*time.Hour {
		t.Errorf("weekly total = %s, want 2h", got)
	}

	monthly, err := aggregateDurations(bs, "month", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}
	if got := monthly["alpha"]["one"]["2026-08"]; got != 2*time.Hour {
		t.Errorf("monthly total = %s, want 2h", got)
	}
}

func TestAggregateDurationsRejectsUnknownTimeframes(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)
	bs := []*block.Block{newBlock(t, "alpha", "one", start, start.Add(time.Hour))}

	if _, err := aggregateDurations(bs, "fortnight", time.Time{}, time.Time{}, nil); errors.Is(
		err, errs.ErrNotATimeframe,
	) == false {
		t.Errorf("aggregateDurations() = %v, want ErrNotATimeframe", err)
	}
}

func TestAggregateDurationsOnAnEmptySet(t *testing.T) {
	stats, err := aggregateDurations(nil, "day", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if got := stats["*"]["*"]["*"]; got != 0 {
		t.Errorf("total = %s, want 0", got)
	}
}

func TestAggregateDurationsCountsTheRunningBlockUpToNow(t *testing.T) {
	running := newBlock(t, "alpha", "one", time.Now().Add(-time.Hour), time.Time{})
	key := running.GetKey()

	stats, err := aggregateDurations(
		[]*block.Block{running}, "day", time.Time{}, time.Time{}, &key)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	got := stats["alpha"]["one"][periodKey(t, "day", running.TimestampStart)]
	if got < 59*time.Minute || got > 61*time.Minute {
		t.Errorf("running block contributed %s, want roughly an hour", got)
	}
}

func TestAggregateDurationsIgnoresAnEndlessBlockThatIsNotRunning(t *testing.T) {
	orphan := newBlock(t, "alpha", "one", time.Now().Add(-time.Hour), time.Time{})

	stats, err := aggregateDurations(
		[]*block.Block{orphan}, "day", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if got := stats["*"]["*"]["*"]; got != 0 {
		t.Errorf("an unfinished block contributed %s, want nothing", got)
	}
}

func TestAggregateDurationsClipsToTheTimeframe(t *testing.T) {
	blockStart := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)
	blockEnd := time.Date(2026, time.August, 26, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		frameStart time.Time
		frameEnd   time.Time
		want       time.Duration
	}{
		{
			name: "unbounded frame keeps the whole block",
			want: 9 * time.Hour,
		},
		{
			name:       "frame inside the block",
			frameStart: blockStart.Add(2 * time.Hour),
			frameEnd:   blockStart.Add(3 * time.Hour),
			want:       time.Hour,
		},
		{
			name:       "frame clips the start",
			frameStart: blockStart.Add(time.Hour),
			frameEnd:   blockEnd.Add(time.Hour),
			want:       8 * time.Hour,
		},
		{
			name:       "frame clips the end",
			frameStart: blockStart.Add(-time.Hour),
			frameEnd:   blockEnd.Add(-4 * time.Hour),
			want:       5 * time.Hour,
		},
		{
			name:       "frame wider than the block",
			frameStart: blockStart.Add(-time.Hour),
			frameEnd:   blockEnd.Add(time.Hour),
			want:       9 * time.Hour,
		},
		{
			name:       "frame entirely before the block",
			frameStart: blockStart.Add(-3 * time.Hour),
			frameEnd:   blockStart.Add(-time.Hour),
			want:       0,
		},
		{
			name:     "open start clips only the end",
			frameEnd: blockStart.Add(time.Hour),
			want:     time.Hour,
		},
		{
			name:       "open end clips only the start",
			frameStart: blockEnd.Add(-time.Hour),
			want:       time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := []*block.Block{newBlock(t, "alpha", "one", blockStart, blockEnd)}

			stats, err := aggregateDurations(
				bs, "day", test.frameStart, test.frameEnd, nil)
			if err != nil {
				t.Fatalf("aggregateDurations() = %s", err)
			}

			if got := stats["*"]["*"]["*"]; got != test.want {
				t.Errorf("total = %s, want %s", got, test.want)
			}
		})
	}
}

func TestAggregateDurationsNeverExceedsTheTimeframe(t *testing.T) {
	frameStart := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
	frameEnd := frameStart.Add(time.Hour)
	frame := frameEnd.Sub(frameStart)

	bs := []*block.Block{
		newBlock(t, "alpha", "one",
			frameStart.Add(-time.Hour), frameStart.Add(30*time.Minute)),
		newBlock(t, "alpha", "two",
			frameStart.Add(30*time.Minute), frameStart.Add(45*time.Minute)),
		newBlock(t, "beta", "one",
			frameStart.Add(45*time.Minute), frameEnd.Add(time.Hour)),
	}

	stats, err := aggregateDurations(bs, "day", frameStart, frameEnd, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if got := stats["*"]["*"]["*"]; got != frame {
		t.Errorf("total = %s, want the full timeframe of %s", got, frame)
	}

	for projectSID, tasks := range stats {
		for taskSID, timeframes := range tasks {
			for key, duration := range timeframes {
				if duration > frame {
					t.Errorf("%s/%s in %s = %s, which exceeds the timeframe of %s",
						projectSID, taskSID, key, duration, frame)
				}
			}
		}
	}
}

func TestAggregateDurationsSumsClippedBlocksWithinTheTimeframe(t *testing.T) {
	frameStart := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
	frameEnd := frameStart.Add(time.Hour)

	bs := []*block.Block{
		newBlock(t, "alpha", "one",
			frameStart.Add(-time.Hour), frameStart.Add(30*time.Minute)),
		newBlock(t, "alpha", "one",
			frameStart.Add(40*time.Minute), frameEnd.Add(time.Hour)),
	}

	stats, err := aggregateDurations(bs, "day", frameStart, frameEnd, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if got := stats["*"]["*"]["*"]; got != 50*time.Minute {
		t.Errorf("total = %s, want 50m of clipped time", got)
	}
}

func TestAggregateDurationsClipsTheRunningBlock(t *testing.T) {
	running := newBlock(t, "alpha", "one", time.Now().Add(-8*time.Hour), time.Time{})
	key := running.GetKey()

	frameStart := time.Now().Add(-time.Hour)
	frameEnd := time.Now().Add(time.Hour)

	stats, err := aggregateDurations(
		[]*block.Block{running}, "day", frameStart, frameEnd, &key)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	got := stats["*"]["*"]["*"]
	if got < 59*time.Minute || got > 61*time.Minute {
		t.Errorf("total = %s, want roughly the hour of the frame it overlaps", got)
	}
}

func TestAggregateDurationsSplitsAcrossDays(t *testing.T) {
	start := time.Date(2026, time.August, 25, 22, 0, 0, 0, time.Local)
	end := time.Date(2026, time.August, 26, 3, 0, 0, 0, time.Local)

	bs := []*block.Block{newBlock(t, "alpha", "night", start, end)}

	stats, err := aggregateDurations(bs, "day", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if got := stats["alpha"]["night"]["2026-08-25"]; got != 2*time.Hour {
		t.Errorf("2026-08-25 = %s, want 2h", got)
	}
	if got := stats["alpha"]["night"]["2026-08-26"]; got != 3*time.Hour {
		t.Errorf("2026-08-26 = %s, want 3h", got)
	}
	if got := stats["*"]["*"]["*"]; got != 5*time.Hour {
		t.Errorf("total = %s, want 5h", got)
	}
}

func TestAggregateDurationsSplitsAcrossManyPeriods(t *testing.T) {
	start := time.Date(2026, time.August, 24, 12, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 4)

	bs := []*block.Block{newBlock(t, "alpha", "long", start, end)}

	stats, err := aggregateDurations(bs, "day", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	days := stats["alpha"]["long"]
	if len(days) != 5 {
		t.Fatalf("the block covers %d days, want 5", len(days))
	}

	if got := days["2026-08-24"]; got != 12*time.Hour {
		t.Errorf("first day = %s, want 12h", got)
	}
	if got := days["2026-08-28"]; got != 12*time.Hour {
		t.Errorf("last day = %s, want 12h", got)
	}
	for _, key := range []string{"2026-08-25", "2026-08-26", "2026-08-27"} {
		if got := days[key]; got != 24*time.Hour {
			t.Errorf("%s = %s, want 24h", key, got)
		}
	}
}

func TestAggregateDurationsSplitsAcrossWeeksAndMonths(t *testing.T) {
	start := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.Local)
	end := time.Date(2026, time.September, 1, 12, 0, 0, 0, time.Local)

	bs := []*block.Block{newBlock(t, "alpha", "turn", start, end)}

	weekly, err := aggregateDurations(bs, "week", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if len(weekly["alpha"]["turn"]) != 2 {
		t.Errorf("the block covers %v, want two weeks", weekly["alpha"]["turn"])
	}
	if got := weekly["alpha"]["turn"]["2026-W35"]; got != 12*time.Hour {
		t.Errorf("2026-W35 = %s, want 12h", got)
	}
	if got := weekly["alpha"]["turn"]["2026-W36"]; got != 36*time.Hour {
		t.Errorf("2026-W36 = %s, want 36h", got)
	}

	monthly, err := aggregateDurations(bs, "month", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if got := monthly["alpha"]["turn"]["2026-08"]; got != 36*time.Hour {
		t.Errorf("2026-08 = %s, want 36h", got)
	}
	if got := monthly["alpha"]["turn"]["2026-09"]; got != 12*time.Hour {
		t.Errorf("2026-09 = %s, want 12h", got)
	}
}

func TestAggregateDurationsOnlyReportsPeriodsInsideTheTimeframe(t *testing.T) {
	start := time.Date(2026, time.August, 25, 22, 0, 0, 0, time.Local)
	end := time.Date(2026, time.August, 26, 3, 0, 0, 0, time.Local)

	frameStart := time.Date(2026, time.August, 26, 0, 0, 0, 0, time.Local)
	frameEnd := time.Date(2026, time.August, 26, 23, 59, 59, 0, time.Local)

	bs := []*block.Block{newBlock(t, "alpha", "night", start, end)}

	stats, err := aggregateDurations(bs, "day", frameStart, frameEnd, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if _, ok := stats["alpha"]["night"]["2026-08-25"]; ok == true {
		t.Errorf("a period outside the timeframe was reported")
	}
	if got := stats["alpha"]["night"]["2026-08-26"]; got != 3*time.Hour {
		t.Errorf("2026-08-26 = %s, want 3h", got)
	}
}

func TestAggregateDurationsSharesSumToTheClippedDuration(t *testing.T) {
	start := time.Date(2026, time.August, 24, 7, 30, 0, 0, time.Local)
	end := time.Date(2026, time.August, 27, 19, 45, 0, 0, time.Local)

	frameStart := time.Date(2026, time.August, 25, 9, 0, 0, 0, time.Local)
	frameEnd := time.Date(2026, time.August, 26, 17, 0, 0, 0, time.Local)

	bs := []*block.Block{newBlock(t, "alpha", "long", start, end)}

	for _, name := range []string{"day", "week", "month"} {
		t.Run(name, func(t *testing.T) {
			stats, err := aggregateDurations(bs, name, frameStart, frameEnd, nil)
			if err != nil {
				t.Fatalf("aggregateDurations() = %s", err)
			}

			var sum time.Duration
			for _, share := range stats["alpha"]["long"] {
				sum += share
			}

			want := timestamp.DurationWithinTimeframe(
				frameStart, frameEnd, start, end)

			if sum != want {
				t.Errorf("the shares sum to %s, want the clipped duration %s",
					sum, want)
			}
			if got := stats["*"]["*"]["*"]; got != want {
				t.Errorf("total = %s, want %s", got, want)
			}
		})
	}
}

func TestAggregateDurationsDropsEmptyPeriods(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.Local)

	bs := []*block.Block{
		newBlock(t, "alpha", "real", start, start.Add(time.Hour)),
		newBlock(t, "alpha", "unfinished", start, time.Time{}),
		newBlock(t, "alpha", "instant", start, start),
	}

	stats, err := aggregateDurations(bs, "day", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	if _, ok := stats["alpha"]["unfinished"]; ok == true {
		t.Errorf("an unfinished block produced a report entry")
	}
	if _, ok := stats["alpha"]["instant"]; ok == true {
		t.Errorf("a zero length block produced a report entry")
	}
	if got := stats["alpha"]["real"][periodKey(t, "day", start)]; got != time.Hour {
		t.Errorf("the real block = %s, want 1h", got)
	}
}

func TestAggregateDurationsSplitsTheRunningBlock(t *testing.T) {
	running := newBlock(t, "alpha", "one", time.Now().Add(-30*time.Minute), time.Time{})
	key := running.GetKey()

	stats, err := aggregateDurations(
		[]*block.Block{running}, "day", time.Time{}, time.Time{}, &key)
	if err != nil {
		t.Fatalf("aggregateDurations() = %s", err)
	}

	var sum time.Duration
	for _, share := range stats["alpha"]["one"] {
		sum += share
	}

	if sum < 29*time.Minute || sum > 31*time.Minute {
		t.Errorf("the running block contributed %s, want roughly 30m", sum)
	}
}
