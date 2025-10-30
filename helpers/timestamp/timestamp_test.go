package timestamp

import (
	"testing"
	"time"
)

type TestDate struct {
	Parse  string
	Result string
}

func TestParse(t *testing.T) {
	var err error
	var tm *Timestamp

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
			Parse:  "-1.5h",
			Result: now.Add(-(90 * time.Minute)).Format(testFmt),
		},
		{
			Parse:  "-0.25h",
			Result: now.Add(-(15 * time.Minute)).Format(testFmt),
		},
		{
			Parse:  "-15m",
			Result: now.Add(-(15 * time.Minute)).Format(testFmt),
		},
		{
			Parse:  "20 minutes ago",
			Result: now.Add(-(20 * time.Minute)).Format(testFmt),
		},
		{
			Parse:  "2 hours ago",
			Result: now.Add(-(2 * time.Hour)).Format(testFmt),
		},
		{
			Parse:  "2 days ago",
			Result: now.Add(-(2 * 24 * time.Hour)).Format(testFmt),
		},
		{
			Parse:  "Yesterday",
			Result: now.Add(-(1 * 24 * time.Hour)).Format(testFmt),
		},
		{
			Parse: "Yesterday 12:00",
			Result: time.Date(now.Year(), now.Month(), now.Day(),
				12, 00, 00, 00, time.Local).Add(-(1 * 24 * time.Hour)).Format(testFmt),
		},
	}

	for _, testdate := range testdates {
		if tm, err = Parse(testdate.Parse); err != nil {
			t.Errorf("Parsing failed: %s\n", err)
			return
		}

		tmf := tm.Time.Format(testFmt)
		if testdate.Result != tmf {
			t.Errorf("Expected '%s', got '%s'\n", testdate.Result, tmf)
			return
		}
		t.Logf("Expected and got '%s'\n", tmf)
	}
}
