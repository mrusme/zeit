package timestamp

import (
	"testing"
	"time"
)

type TestDate struct {
	Parse  string
	Result string
}

type TestOffset struct {
	Parse  string
	Offset time.Duration
}

func TestParse(t *testing.T) {
	testFmt := "2006-01-02 15:04:05 -0700"

	now := time.Now()

	testdates := []TestDate{
		{
			Parse: "today 12:00",
			Result: time.Date(now.Year(), now.Month(), now.Day(),
				12, 00, 00, 00, time.Local).Format(testFmt),
		},
		{
			Parse: "16.9.2025 12:00",
			Result: time.Date(2025, 9, 16,
				12, 00, 00, 00, time.Local).Format(testFmt),
		},
		{
			Parse: "9/16/2025 12:00",
			Result: time.Date(2025, 9, 16,
				12, 00, 00, 00, time.Local).Format(testFmt),
		},
		{
			Parse: "Yesterday 12:00",
			Result: time.Date(now.Year(), now.Month(), now.Day(),
				12, 00, 00, 00, time.Local).Add(-(1 * 24 * time.Hour)).Format(testFmt),
		},
	}

	for _, testdate := range testdates {
		t.Run(testdate.Parse, func(t *testing.T) {
			tm, err := Parse(testdate.Parse)
			if err != nil {
				t.Fatalf("Parsing failed: %s", err)
			}

			tmf := tm.Time.Format(testFmt)
			if testdate.Result != tmf {
				t.Errorf("Expected '%s', got '%s'", testdate.Result, tmf)
			}
		})
	}
}

func TestParseRelative(t *testing.T) {
	testoffsets := []TestOffset{
		{Parse: "-1.5h", Offset: -(90 * time.Minute)},
		{Parse: "-0.25h", Offset: -(15 * time.Minute)},
		{Parse: "-15m", Offset: -(15 * time.Minute)},
		{Parse: "20 minutes ago", Offset: -(20 * time.Minute)},
		{Parse: "2 hours ago", Offset: -(2 * time.Hour)},
		{Parse: "2 days ago", Offset: -(2 * 24 * time.Hour)},
		{Parse: "Yesterday", Offset: -(1 * 24 * time.Hour)},
	}

	for _, testoffset := range testoffsets {
		t.Run(testoffset.Parse, func(t *testing.T) {
			before := time.Now()
			tm, err := Parse(testoffset.Parse)
			after := time.Now()

			if err != nil {
				t.Fatalf("Parsing failed: %s", err)
			}

			earliest := before.Add(testoffset.Offset)
			latest := after.Add(testoffset.Offset)

			if tm.Time.Before(earliest) == true || tm.Time.After(latest) == true {
				t.Errorf("Expected a time between '%s' and '%s', got '%s'",
					earliest, latest, tm.Time)
			}
		})
	}
}
