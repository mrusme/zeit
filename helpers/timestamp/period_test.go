package timestamp

import (
	"errors"
	"testing"
	"time"

	"xn--gckvb8fzb.com/zeit/errs"
)

func withLocation(t *testing.T, name string) {
	t.Helper()

	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Skipf("Timezone %s unavailable: %s", name, err)
	}

	previous := time.Local
	time.Local = loc
	t.Cleanup(func() {
		time.Local = previous
	})
}

func TestParsePeriodAcceptsBareAndFramedPeriods(t *testing.T) {
	periods := []string{"hour", "day", "week", "month", "quarter", "year"}
	frames := []string{"", "this ", "current ", "last ", "previous "}

	for _, period := range periods {
		for _, frame := range frames {
			input := frame + period

			t.Run(input, func(t *testing.T) {
				ts, err := ParsePeriod(input)
				if err != nil {
					t.Fatalf("ParsePeriod(%q) = %s", input, err)
				}

				if ts.IsRange == false {
					t.Errorf("ParsePeriod(%q).IsRange = false, want true", input)
				}
				if ts.Time.IsZero() == true {
					t.Errorf("ParsePeriod(%q).Time is zero", input)
				}
				if ts.ToTime.After(ts.Time) == false {
					t.Errorf("ParsePeriod(%q) = %s -> %s, want an increasing range",
						input, ts.Time, ts.ToTime)
				}
			})
		}
	}
}

func TestParsePeriodRejectsUnknownInput(t *testing.T) {
	inputs := []string{
		"",
		"not a period",
		"thisweek",
		"next week",
		"week ago",
		"2 weeks",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if _, err := ParsePeriod(input); err == nil {
				t.Errorf("ParsePeriod(%q) = nil error, want an error", input)
			}
		})
	}
}

func TestParsePeriodUnhandledPeriodsReturnAnEmptyRange(t *testing.T) {
	for _, input := range []string{"decade", "century"} {
		t.Run(input, func(t *testing.T) {
			ts, err := ParsePeriod(input)
			if err != nil {
				t.Fatalf("ParsePeriod(%q) = %s", input, err)
			}

			if ts.Time.IsZero() == false || ts.ToTime.IsZero() == false {
				t.Errorf("ParsePeriod(%q) = %s -> %s, want a zero range",
					input, ts.Time, ts.ToTime)
			}
		})
	}
}

func TestParsePeriodBareFormMatchesThisForm(t *testing.T) {
	for _, period := range []string{"day", "week", "month", "quarter", "year"} {
		t.Run(period, func(t *testing.T) {
			bare, err := ParsePeriod(period)
			if err != nil {
				t.Fatalf("ParsePeriod(%q) = %s", period, err)
			}

			framed, err := ParsePeriod("this " + period)
			if err != nil {
				t.Fatalf("ParsePeriod(%q) = %s", "this "+period, err)
			}

			if bare.Time.Equal(framed.Time) == false ||
				bare.ToTime.Equal(framed.ToTime) == false {
				t.Errorf("ParsePeriod(%q) = %s -> %s, ParsePeriod(%q) = %s -> %s",
					period, bare.Time, bare.ToTime,
					"this "+period, framed.Time, framed.ToTime)
			}
		})
	}
}

func TestParsePeriodStartsAtLocalMidnight(t *testing.T) {
	zones := []string{
		"UTC",
		"Europe/Berlin",
		"America/Chicago",
		"Asia/Tokyo",
		"Australia/Adelaide",
	}

	for _, zone := range zones {
		for _, period := range []string{"day", "week", "month", "quarter", "year"} {
			t.Run(zone+" "+period, func(t *testing.T) {
				withLocation(t, zone)

				ts, err := ParsePeriod(period)
				if err != nil {
					t.Fatalf("ParsePeriod(%q) = %s", period, err)
				}

				h, m, s := ts.Time.Clock()
				if h != 0 || m != 0 || s != 0 {
					t.Errorf("ParsePeriod(%q) starts at %s, want local midnight",
						period, ts.Time)
				}

				h, m, s = ts.ToTime.Clock()
				if h != 23 || m != 59 || s != 59 {
					t.Errorf("ParsePeriod(%q) ends at %s, want 23:59:59 local",
						period, ts.ToTime)
				}
			})
		}
	}
}

func TestParsePeriodWeekRunsMondayToSunday(t *testing.T) {
	for _, zone := range []string{"UTC", "Europe/Berlin", "America/Chicago"} {
		t.Run(zone, func(t *testing.T) {
			withLocation(t, zone)

			ts, err := ParsePeriod("this week")
			if err != nil {
				t.Fatalf("ParsePeriod() = %s", err)
			}

			if ts.Time.Weekday() != time.Monday {
				t.Errorf("week starts on %s, want Monday", ts.Time.Weekday())
			}
			if ts.ToTime.Weekday() != time.Sunday {
				t.Errorf("week ends on %s, want Sunday", ts.ToTime.Weekday())
			}

			now := time.Now()
			if now.Before(ts.Time) || now.After(ts.ToTime) {
				t.Errorf("this week = %s -> %s, which excludes now (%s)",
					ts.Time, ts.ToTime, now)
			}
		})
	}
}

func TestParsePeriodLastWeekPrecedesThisWeek(t *testing.T) {
	this, err := ParsePeriod("this week")
	if err != nil {
		t.Fatalf("ParsePeriod() = %s", err)
	}

	last, err := ParsePeriod("last week")
	if err != nil {
		t.Fatalf("ParsePeriod() = %s", err)
	}

	if last.ToTime.After(this.Time) == true {
		t.Errorf("last week ends at %s, which is not before this week's start %s",
			last.ToTime, this.Time)
	}

	gap := this.Time.Sub(last.ToTime)
	if gap < 0 || gap > time.Second {
		t.Errorf("gap between last and this week is %s, want at most one second", gap)
	}
}

func TestParsePeriodContainsNow(t *testing.T) {
	for _, period := range []string{"hour", "day", "week", "month", "quarter", "year"} {
		t.Run(period, func(t *testing.T) {
			ts, err := ParsePeriod("this " + period)
			if err != nil {
				t.Fatalf("ParsePeriod() = %s", err)
			}

			now := time.Now()
			if now.Before(ts.Time) || now.After(ts.ToTime) {
				t.Errorf("this %s = %s -> %s, which excludes now (%s)",
					period, ts.Time, ts.ToTime, now)
			}
		})
	}
}

func TestQuarterPeriod(t *testing.T) {
	withLocation(t, "Europe/Berlin")

	period, err := GetPeriod("quarter")
	if err != nil {
		t.Fatalf("GetPeriod() = %s", err)
	}

	tests := []struct {
		name      string
		now       time.Time
		previous  bool
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "february this quarter",
			now:       time.Date(2026, time.February, 15, 12, 0, 0, 0, time.Local),
			wantStart: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2026, time.March, 31, 23, 59, 59, 0, time.Local),
		},
		{
			name:      "february last quarter wraps the year",
			now:       time.Date(2026, time.February, 15, 12, 0, 0, 0, time.Local),
			previous:  true,
			wantStart: time.Date(2025, time.October, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2025, time.December, 31, 23, 59, 59, 0, time.Local),
		},
		{
			name:      "january last quarter wraps the year",
			now:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local),
			previous:  true,
			wantStart: time.Date(2025, time.October, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2025, time.December, 31, 23, 59, 59, 0, time.Local),
		},
		{
			name:      "august this quarter",
			now:       time.Date(2026, time.August, 26, 12, 0, 0, 0, time.Local),
			wantStart: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2026, time.September, 30, 23, 59, 59, 0, time.Local),
		},
		{
			name:      "august last quarter",
			now:       time.Date(2026, time.August, 26, 12, 0, 0, 0, time.Local),
			previous:  true,
			wantStart: time.Date(2026, time.April, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2026, time.June, 30, 23, 59, 59, 0, time.Local),
		},
		{
			name:      "december this quarter",
			now:       time.Date(2026, time.December, 31, 23, 0, 0, 0, time.Local),
			wantStart: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2026, time.December, 31, 23, 59, 59, 0, time.Local),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := period.Start(test.now)
			if test.previous == true {
				start = period.Previous(start)
			}

			end := period.Next(start).Add(-time.Second)

			if start.Equal(test.wantStart) == false {
				t.Errorf("start = %s, want %s", start, test.wantStart)
			}
			if end.Equal(test.wantEnd) == false {
				t.Errorf("end = %s, want %s", end, test.wantEnd)
			}
		})
	}
}

func TestParsePeriodSpansDaylightSavingChanges(t *testing.T) {
	withLocation(t, "Europe/Berlin")

	ts, err := ParsePeriod("this week")
	if err != nil {
		t.Fatalf("ParsePeriod() = %s", err)
	}

	if ts.Time.Weekday() != time.Monday || ts.ToTime.Weekday() != time.Sunday {
		t.Errorf("week = %s -> %s, want Monday to Sunday", ts.Time, ts.ToTime)
	}

	span := ts.ToTime.Sub(ts.Time)
	if span < 6*24*time.Hour || span > 8*24*time.Hour {
		t.Errorf("week spans %s, want roughly seven days", span)
	}
}

func TestParseFallsBackToDateParser(t *testing.T) {
	ts, err := Parse("this week")
	if err != nil {
		t.Fatalf("Parse() = %s", err)
	}
	if ts.IsRange == false {
		t.Errorf("Parse(\"this week\").IsRange = false, want true")
	}

	ts, err = Parse("2 hours ago")
	if err != nil {
		t.Fatalf("Parse() = %s", err)
	}
	if ts.IsRange == true {
		t.Errorf("Parse(\"2 hours ago\").IsRange = true, want false")
	}
	if ts.Time.IsZero() == true {
		t.Errorf("Parse(\"2 hours ago\").Time is zero")
	}

	if _, err = Parse("definitely not a date"); err == nil {
		t.Errorf("Parse() = nil error, want an error")
	}
}

func TestGetPeriodRejectsUnknownNames(t *testing.T) {
	for _, name := range []string{"", "fortnight", "decade", "century"} {
		if _, err := GetPeriod(name); errors.Is(err, errs.ErrNotATimeframe) == false {
			t.Errorf("GetPeriod(%q) = %v, want ErrNotATimeframe", name, err)
		}
	}
}

func TestGetPeriodIsCaseInsensitive(t *testing.T) {
	for _, name := range []string{"day", "Day", "DAY"} {
		if _, err := GetPeriod(name); err != nil {
			t.Errorf("GetPeriod(%q) = %s", name, err)
		}
	}
}

func TestPeriodKeys(t *testing.T) {
	at := time.Date(2026, time.August, 26, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		want string
	}{
		{name: "hour", want: "2026-08-26T14"},
		{name: "day", want: "2026-08-26"},
		{name: "week", want: "2026-W35"},
		{name: "month", want: "2026-08"},
		{name: "quarter", want: "2026-Q3"},
		{name: "year", want: "2026"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			period, err := GetPeriod(test.name)
			if err != nil {
				t.Fatalf("GetPeriod() = %s", err)
			}

			if got := period.Key(at); got != test.want {
				t.Errorf("Key() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWeekKeyUsesTheISOYear(t *testing.T) {
	period, err := GetPeriod("week")
	if err != nil {
		t.Fatalf("GetPeriod() = %s", err)
	}

	at := time.Date(2027, time.January, 1, 12, 0, 0, 0, time.UTC)

	if isoYear, _ := at.ISOWeek(); isoYear == at.Year() {
		t.Fatalf("%s no longer straddles an ISO year boundary", at)
	}

	if got := period.Key(at); got != "2026-W53" {
		t.Errorf("Key() = %q, want 2026-W53", got)
	}
}

func TestWeekKeyKeepsOneISOWeekUnderOneKey(t *testing.T) {
	period, err := GetPeriod("week")
	if err != nil {
		t.Fatalf("GetPeriod() = %s", err)
	}

	december := time.Date(2026, time.December, 31, 12, 0, 0, 0, time.UTC)
	january := time.Date(2027, time.January, 1, 12, 0, 0, 0, time.UTC)

	decemberYear, decemberWeek := december.ISOWeek()
	januaryYear, januaryWeek := january.ISOWeek()

	if decemberYear != januaryYear || decemberWeek != januaryWeek {
		t.Fatalf("the two dates are no longer in the same ISO week")
	}

	if period.Key(december) != period.Key(january) {
		t.Errorf("Key() = %q and %q for two dates in the same ISO week",
			period.Key(december), period.Key(january))
	}
}

func TestWeekKeysSortChronologicallyAsStrings(t *testing.T) {
	period, err := GetPeriod("week")
	if err != nil {
		t.Fatalf("GetPeriod() = %s", err)
	}

	var keys []string

	at := time.Date(2026, time.January, 5, 12, 0, 0, 0, time.UTC)
	for range 53 {
		keys = append(keys, period.Key(at))
		at = period.Next(at)
	}

	for i := 1; i < len(keys); i++ {
		if keys[i-1] >= keys[i] {
			t.Errorf("%q does not sort before %q", keys[i-1], keys[i])
		}
	}
}

func TestPeriodKeysSortChronologicallyAsStrings(t *testing.T) {
	for _, name := range []string{"hour", "day", "week", "month", "year"} {
		t.Run(name, func(t *testing.T) {
			period, err := GetPeriod(name)
			if err != nil {
				t.Fatalf("GetPeriod() = %s", err)
			}

			at := period.Start(
				time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC),
			)

			previous := period.Key(at)
			for range 40 {
				at = period.Next(at)

				if current := period.Key(at); current <= previous {
					t.Errorf("%q does not sort after %q", current, previous)
				} else {
					previous = current
				}
			}
		})
	}
}

func TestPeriodStartIsIdempotent(t *testing.T) {
	withLocation(t, "Europe/Berlin")

	at := time.Date(2026, time.August, 26, 14, 37, 12, 500, time.Local)

	for _, name := range []string{"hour", "day", "week", "month", "quarter", "year"} {
		t.Run(name, func(t *testing.T) {
			period, err := GetPeriod(name)
			if err != nil {
				t.Fatalf("GetPeriod() = %s", err)
			}

			start := period.Start(at)

			if period.Start(start).Equal(start) == false {
				t.Errorf("Start(Start(t)) = %s, want %s", period.Start(start), start)
			}
			if start.After(at) == true {
				t.Errorf("Start() = %s, which is after %s", start, at)
			}
			if period.Next(start).After(at) == false {
				t.Errorf("Next(Start()) = %s, which does not contain %s",
					period.Next(start), at)
			}
		})
	}
}

func TestPeriodNextAndPreviousAreInverse(t *testing.T) {
	withLocation(t, "Europe/Berlin")

	at := time.Date(2026, time.March, 30, 0, 0, 0, 0, time.Local)

	for _, name := range []string{"hour", "day", "week", "month", "quarter", "year"} {
		t.Run(name, func(t *testing.T) {
			period, err := GetPeriod(name)
			if err != nil {
				t.Fatalf("GetPeriod() = %s", err)
			}

			start := period.Start(at)

			if period.Previous(period.Next(start)).Equal(start) == false {
				t.Errorf("Previous(Next(t)) = %s, want %s",
					period.Previous(period.Next(start)), start)
			}
		})
	}
}

func TestDayPeriodSpansDaylightSavingChanges(t *testing.T) {
	withLocation(t, "Europe/Berlin")

	period, err := GetPeriod("day")
	if err != nil {
		t.Fatalf("GetPeriod() = %s", err)
	}

	tests := []struct {
		name string
		date time.Time
		want time.Duration
	}{
		{
			name: "spring forward",
			date: time.Date(2026, time.March, 29, 12, 0, 0, 0, time.Local),
			want: 23 * time.Hour,
		},
		{
			name: "fall back",
			date: time.Date(2026, time.October, 25, 12, 0, 0, 0, time.Local),
			want: 25 * time.Hour,
		},
		{
			name: "ordinary day",
			date: time.Date(2026, time.August, 26, 12, 0, 0, 0, time.Local),
			want: 24 * time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := period.Start(test.date)

			if got := period.Next(start).Sub(start); got != test.want {
				t.Errorf("the day spans %s, want %s", got, test.want)
			}
			if h, m, s := period.Next(start).Clock(); h != 0 || m != 0 || s != 0 {
				t.Errorf("the next day starts at %s, want local midnight",
					period.Next(start))
			}
		})
	}
}
