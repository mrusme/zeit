package timestamp

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
)

type Timestamp struct {
	Time    time.Time
	ToTime  time.Time
	IsRange bool
}

var periodRegex = regexp.MustCompile(`(?m)(this|current|last|previous){0,1}\s*(week|month|quarter|year|decade|century)`)

func ParsePeriod(str string) (*Timestamp, error) {
	var frame string
	var period string
	var now time.Time = time.Now()

	ts := new(Timestamp)

	matches := periodRegex.FindStringSubmatch(str)

	if len(matches) == 2 {
		period = strings.ToLower(matches[1])
	} else if len(matches) == 3 {
		frame = strings.ToLower(matches[1])
		period = strings.ToLower(matches[2])
	} else {
		return nil, errors.New("No period found")
	}

	ts.IsRange = true

	previousPeriod := false
	if frame == "last" || frame == "previous" {
		previousPeriod = true
	}

	switch period {
	case "week":
		weekday := now.Weekday()
		daysToMonday := (weekday - time.Monday + 7) % 7
		ts.Time = now.
			Add(-time.Duration(daysToMonday) * 24 * time.Hour).
			Truncate(24 * time.Hour)
		if previousPeriod {
			ts.Time = ts.Time.
				Add(-7 * 24 * time.Hour)
		}
		ts.ToTime = ts.Time.
			Add(6 * 24 * time.Hour).
			Add(23 * time.Hour).
			Add(59 * time.Minute).
			Add(59 * time.Second)
	case "month":
		if previousPeriod == false {
			ts.Time = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		} else {
			ts.Time = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		}
		nextMonth := ts.Time.AddDate(0, 1, 0)
		ts.ToTime = nextMonth.Add(-time.Second)
	case "year":
		if previousPeriod == false {
			ts.Time = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
			ts.ToTime = time.Date(now.Year(), time.December, 31, 23, 59, 59, 0, now.Location())
		} else {
			ts.Time = time.Date(now.Year()-1, time.January, 1, 0, 0, 0, 0, now.Location())
			ts.ToTime = time.Date(now.Year()-1, time.December, 31, 23, 59, 59, 0, now.Location())
		}
	}

	return ts, nil
}

func Parse(str string) (*Timestamp, error) {
	var err error
	var dt date.Date

	var ts *Timestamp
	ts, err = ParsePeriod(str)
	if err == nil {
		return ts, nil
	} else {
		ts = new(Timestamp)
	}

	cfg := dateparser.Configuration{
		DefaultTimezone: time.Local,
	}

	if dt, err = dateparser.Parse(&cfg, str); err != nil {
		return nil, err
	}

	ts.Time = dt.Time

	return ts, nil
}
