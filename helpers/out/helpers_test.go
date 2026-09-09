package out

import (
	"encoding/json"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestEndTimestampMarshalsMissingEndsAsNull(t *testing.T) {
	got, err := json.Marshal(EndTimestamp(time.Time{}))
	if err != nil {
		t.Fatalf("json.Marshal() = %s", err)
	}

	if string(got) != "null" {
		t.Errorf("json.Marshal() = %s, want null", got)
	}
}

func TestEndTimestampMarshalsRealTimestamps(t *testing.T) {
	ts := time.Date(2026, time.August, 26, 17, 30, 0, 0, time.UTC)

	got, err := json.Marshal(EndTimestamp(ts))
	if err != nil {
		t.Fatalf("json.Marshal() = %s", err)
	}

	var back time.Time
	if err = json.Unmarshal(got, &back); err != nil {
		t.Fatalf("json.Unmarshal(%s) = %s", got, err)
	}

	if back.Equal(ts) == false {
		t.Errorf("round trip = %s, want %s", back, ts)
	}
}

func TestEndTimestampString(t *testing.T) {
	ts := time.Date(2026, time.August, 26, 17, 30, 0, 0, time.UTC)

	if got := EndTimestamp(ts).String(); got != "2026-08-26 17:30:00" {
		t.Errorf("String() = %q, want the DateTime layout", got)
	}

	if got := EndTimestamp(time.Time{}).String(); got != "not ended" {
		t.Errorf("String() = %q, want a placeholder", got)
	}
}

func TestSecondsMarshalsAsWholeSeconds(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{name: "zero", d: 0, want: "0"},
		{name: "one second", d: time.Second, want: "1"},
		{name: "two hours", d: 2 * time.Hour, want: "7200"},
		{name: "sub second truncates", d: 900 * time.Millisecond, want: "0"},
		{name: "rounds toward zero", d: 1900 * time.Millisecond, want: "1"},
		{name: "negative", d: -time.Second, want: "-1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := json.Marshal(Seconds(test.d))
			if err != nil {
				t.Fatalf("json.Marshal() = %s", err)
			}

			if string(got) != test.want {
				t.Errorf("json.Marshal(%s) = %s, want %s", test.d, got, test.want)
			}
		})
	}
}

func TestSecondsString(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{d: 0, want: "0s"},
		{d: 90 * time.Second, want: "1m30s"},
		{d: 2 * time.Hour, want: "2h0m0s"},
		{d: 1499 * time.Millisecond, want: "1s"},
	}

	for _, test := range tests {
		if got := Seconds(test.d).String(); got != test.want {
			t.Errorf("Seconds(%s).String() = %q, want %q", test.d, got, test.want)
		}
	}
}

func TestStatusOutAlwaysCarriesEveryKey(t *testing.T) {
	got, err := json.Marshal(new(StatusOut))
	if err != nil {
		t.Fatalf("json.Marshal() = %s", err)
	}

	var decoded map[string]any
	if err = json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() = %s", err)
	}

	for _, key := range []string{
		"status", "is_running", "project_sid", "task_sid", "timer",
	} {
		if _, ok := decoded[key]; ok == false {
			t.Errorf("key %q is missing from %s", key, got)
		}
	}
}

func TestRandomVisibleHexColor(t *testing.T) {
	pattern := regexp.MustCompile(`^#[0-9A-F]{6}$`)

	for range 64 {
		color := RandomVisibleHexColor()

		if pattern.MatchString(color) == false {
			t.Fatalf("RandomVisibleHexColor() = %q, want a hex color", color)
		}

		for i := 1; i < 7; i += 2 {
			component, err := strconv.ParseInt(color[i:i+2], 16, 0)
			if err != nil {
				t.Fatalf("parsing %q: %s", color, err)
			}

			if component < 64 || component > 191 {
				t.Errorf("RandomVisibleHexColor() = %q, component %d is outside 64-191",
					color, component)
			}
		}
	}
}

func TestColorAcceptsHexStrings(t *testing.T) {
	if Color("#AABBCC") == nil {
		t.Errorf("Color() = nil for a valid hex color")
	}
}
