package timestamp

import (
	"testing"
	"time"
)

var (
	frameStart = time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)
	frameEnd   = time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	before     = time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	inside     = time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	after      = time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
)

func TestIsWithinTimeframe(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		value time.Time
		want  bool
	}{
		{name: "inside", start: frameStart, end: frameEnd, value: inside, want: true},
		{name: "before", start: frameStart, end: frameEnd, value: before, want: false},
		{name: "after", start: frameStart, end: frameEnd, value: after, want: false},
		{
			name:  "on the lower bound",
			start: frameStart, end: frameEnd, value: frameStart, want: true,
		},
		{
			name:  "on the upper bound",
			start: frameStart, end: frameEnd, value: frameEnd, want: true,
		},
		{
			name:  "unbounded frame accepts anything",
			start: time.Time{}, end: time.Time{}, value: before, want: true,
		},
		{
			name:  "open start accepts earlier values",
			start: time.Time{}, end: frameEnd, value: before, want: true,
		},
		{
			name:  "open start still rejects later values",
			start: time.Time{}, end: frameEnd, value: after, want: false,
		},
		{
			name:  "open end accepts later values",
			start: frameStart, end: time.Time{}, value: after, want: true,
		},
		{
			name:  "open end still rejects earlier values",
			start: frameStart, end: time.Time{}, value: before, want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := IsWithinTimeframe(test.start, test.end, test.value)
			if got != test.want {
				t.Errorf("IsWithinTimeframe() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestIsEndWithinTimeframeTreatsMissingEndsAsNow(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	if IsEndWithinTimeframe(past, future, time.Time{}) == false {
		t.Errorf("a running block is not inside a frame that contains now")
	}

	if IsEndWithinTimeframe(
		past.Add(-48*time.Hour), past, time.Time{},
	) == true {
		t.Errorf("a running block is inside a frame that has already closed")
	}
}

func TestIsFullyWithinTimeframe(t *testing.T) {
	tests := []struct {
		name string
		vs   time.Time
		ve   time.Time
		want bool
	}{
		{name: "both inside", vs: inside, ve: inside.Add(time.Hour), want: true},
		{name: "starts before", vs: before, ve: inside, want: false},
		{name: "ends after", vs: inside, ve: after, want: false},
		{name: "spans the frame", vs: before, ve: after, want: false},
		{name: "entirely before", vs: before, ve: before.Add(time.Hour), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := IsFullyWithinTimeframe(frameStart, frameEnd, test.vs, test.ve)
			if got != test.want {
				t.Errorf("IsFullyWithinTimeframe() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestIsPartiallyWithinTimeframe(t *testing.T) {
	tests := []struct {
		name string
		vs   time.Time
		ve   time.Time
		want bool
	}{
		{name: "both inside", vs: inside, ve: inside.Add(time.Hour), want: true},
		{name: "starts before and ends inside", vs: before, ve: inside, want: true},
		{name: "starts inside and ends after", vs: inside, ve: after, want: true},
		{name: "spans the whole frame", vs: before, ve: after, want: true},
		{name: "entirely before", vs: before, ve: before.Add(time.Hour), want: false},
		{name: "entirely after", vs: after, ve: after.Add(time.Hour), want: false},
		{name: "ends on the lower bound", vs: before, ve: frameStart, want: true},
		{name: "starts on the upper bound", vs: frameEnd, ve: after, want: true},
		{
			name: "ends just before the lower bound",
			vs:   before,
			ve:   frameStart.Add(-time.Nanosecond),
			want: false,
		},
		{
			name: "starts just after the upper bound",
			vs:   frameEnd.Add(time.Nanosecond),
			ve:   after,
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := IsPartiallyWithinTimeframe(frameStart, frameEnd, test.vs, test.ve)
			if got != test.want {
				t.Errorf("IsPartiallyWithinTimeframe() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestIsPartiallyWithinUnboundedTimeframe(t *testing.T) {
	values := [][2]time.Time{
		{before, before.Add(time.Hour)},
		{inside, inside.Add(time.Hour)},
		{after, after.Add(time.Hour)},
		{inside, time.Time{}},
	}

	for _, value := range values {
		if IsPartiallyWithinTimeframe(
			time.Time{}, time.Time{}, value[0], value[1],
		) == false {
			t.Errorf("an unbounded timeframe excluded %s -> %s", value[0], value[1])
		}
	}
}

func TestDurationWithinTimeframe(t *testing.T) {
	vStart := time.Date(2026, time.August, 15, 9, 0, 0, 0, time.UTC)
	vEnd := time.Date(2026, time.August, 15, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  time.Duration
	}{
		{name: "unbounded frame", want: 9 * time.Hour},
		{
			name:  "frame inside the value",
			start: vStart.Add(2 * time.Hour),
			end:   vStart.Add(3 * time.Hour),
			want:  time.Hour,
		},
		{
			name:  "frame wider than the value",
			start: vStart.Add(-time.Hour),
			end:   vEnd.Add(time.Hour),
			want:  9 * time.Hour,
		},
		{
			name:  "frame clips the start",
			start: vStart.Add(4 * time.Hour),
			end:   vEnd.Add(time.Hour),
			want:  5 * time.Hour,
		},
		{
			name:  "frame clips the end",
			start: vStart.Add(-time.Hour),
			end:   vEnd.Add(-3 * time.Hour),
			want:  6 * time.Hour,
		},
		{
			name:  "frame entirely before",
			start: vStart.Add(-3 * time.Hour),
			end:   vStart.Add(-time.Hour),
			want:  0,
		},
		{
			name:  "frame entirely after",
			start: vEnd.Add(time.Hour),
			end:   vEnd.Add(3 * time.Hour),
			want:  0,
		},
		{
			name:  "frame touching the start",
			start: vStart.Add(-time.Hour),
			end:   vStart,
			want:  0,
		},
		{
			name:  "frame touching the end",
			start: vEnd,
			end:   vEnd.Add(time.Hour),
			want:  0,
		},
		{
			name: "open start clips only the end",
			end:  vStart.Add(time.Hour),
			want: time.Hour,
		},
		{
			name:  "open end clips only the start",
			start: vEnd.Add(-time.Hour),
			want:  time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := DurationWithinTimeframe(test.start, test.end, vStart, vEnd)
			if got != test.want {
				t.Errorf("DurationWithinTimeframe() = %s, want %s", got, test.want)
			}
		})
	}
}

func TestDurationWithinTimeframeNeverExceedsTheFrame(t *testing.T) {
	start := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	frame := end.Sub(start)

	values := [][2]time.Time{
		{start.Add(-100 * time.Hour), end.Add(100 * time.Hour)},
		{start, end},
		{start.Add(-time.Minute), end.Add(time.Minute)},
		{start.Add(30 * time.Minute), end.Add(30 * time.Minute)},
	}

	for _, value := range values {
		if got := DurationWithinTimeframe(
			start, end, value[0], value[1],
		); got > frame {
			t.Errorf("DurationWithinTimeframe() = %s for %s -> %s, which exceeds %s",
				got, value[0], value[1], frame)
		}
	}
}

func TestDurationWithinTimeframeWithoutAnEnd(t *testing.T) {
	vStart := time.Date(2026, time.August, 15, 9, 0, 0, 0, time.UTC)

	if got := DurationWithinTimeframe(
		time.Time{}, time.Time{}, vStart, time.Time{},
	); got != 0 {
		t.Errorf("DurationWithinTimeframe() = %s for a value with no end, want 0", got)
	}
}

func TestDurationWithinTimeframeWithAnInvertedValue(t *testing.T) {
	vStart := time.Date(2026, time.August, 15, 18, 0, 0, 0, time.UTC)
	vEnd := time.Date(2026, time.August, 15, 9, 0, 0, 0, time.UTC)

	if got := DurationWithinTimeframe(
		time.Time{}, time.Time{}, vStart, vEnd,
	); got != 0 {
		t.Errorf("DurationWithinTimeframe() = %s for an inverted value, want 0", got)
	}
}

func TestIsPartiallyWithinTimeframeIncludesSpanningValues(t *testing.T) {
	if IsPartiallyWithinTimeframe(frameStart, frameEnd, before, after) == false {
		t.Errorf("a value that starts before and ends after the frame is " +
			"reported as outside it")
	}
}

func TestIsPartiallyWithinTimeframeIncludesARunningSpanningValue(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour)
	frameStart := time.Now().Add(-time.Hour)
	frameEnd := time.Now().Add(time.Hour)

	if IsPartiallyWithinTimeframe(
		frameStart, frameEnd, past, time.Time{},
	) == false {
		t.Errorf("a still running value that started before the frame is " +
			"reported as outside it")
	}
}

func TestIsPartiallyWithinTimeframeIsSymmetricWithTheFrame(t *testing.T) {
	values := [][2]time.Time{
		{before, after},
		{before, inside},
		{inside, after},
		{inside, inside.Add(time.Hour)},
	}

	for _, value := range values {
		direct := IsPartiallyWithinTimeframe(frameStart, frameEnd, value[0], value[1])
		flipped := IsPartiallyWithinTimeframe(value[0], value[1], frameStart, frameEnd)

		if direct != flipped {
			t.Errorf("overlap of %s -> %s with the frame is %t one way and %t the other",
				value[0], value[1], direct, flipped)
		}
	}
}
