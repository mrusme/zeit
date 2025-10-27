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

var periodRegex = regexp.MustCompile(
	`(?m)^(this|current|last|previous){0,1}\s+(hour|day|week|month|quarter|year|decade|century)$`,
)

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
	case "hour":
		hour := now.Hour()
		if previousPeriod {
			hour -= 1
		}
		ts.Time = time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
		ts.ToTime = ts.Time.Add(59 * time.Minute).Add(59 * time.Second)
	case "day":
		day := now.Day()
		if previousPeriod {
			day -= 1
		}
		ts.Time = time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
		ts.ToTime = ts.Time.Add(23 * time.Hour).Add(59 * time.Minute).Add(59 * time.Second)
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
	case "quarter":
		ts.Time, ts.ToTime = getQuarterStartEnd(now, previousPeriod)
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

func getQuarterStartEnd(now time.Time, last bool) (time.Time, time.Time) {
	month := int(now.Month())
	var quarterStartMonth int
	var quarterEndMonth int

	if last {
		// If 'last' is true, work with the previous quarter
		month -= 3
		if month <= 0 {
			month += 12
		}
	}

	// "Wait, wat, what is this black sorcery?" you might be asking yourself.
	// If you type e.g. (9-1)/3*3+1 into your calculator you'll be getting 9.
	// However, if you run this calculation in Go, you'll be getting 7.
	//
	// The reason for this is Go's way of handling integer calculations when
	// floats are involved. The formula used here ( (9-1)/3*3+1 ) could be
	// (more transparently) expressed using float values like so:
	//
	// math.Floor((9.0-1.0)/3.0)*3.0 + 1.0
	//
	// This would return the desired result of 7(.0). However, by using integers
	// we're saving ourselves having to explicitly pull in the math package and
	// call the Floor function.
	quarterStartMonth = (month-1)/3*3 + 1
	quarterEndMonth = quarterStartMonth + 2

	qStart := time.Date(now.Year(), time.Month(quarterStartMonth), 1, 0, 0, 0, 0, time.UTC)
	qEnd := time.Date(now.Year(), time.Month(quarterEndMonth+1), 0, 23, 59, 59, 0, time.UTC)

	return qStart, qEnd
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

func IsStartWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
) bool {
	if timeframeStart.IsZero() == false &&
		(vStart.Before(timeframeStart) || vStart.After(timeframeEnd)) {
		return false
	}

	return true
}

func IsEndWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vEnd time.Time,
) bool {
	if timeframeEnd.IsZero() == false &&
		(vEnd.Before(timeframeStart) || vEnd.After(timeframeEnd)) {
		return false
	}

	return true
}

func IsFullyWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
	vEnd time.Time,
) bool {
	if IsStartWithinTimeframe(
		timeframeStart, timeframeEnd, vStart,
	) == false {
		return false
	}
	if IsEndWithinTimeframe(
		timeframeStart, timeframeEnd, vEnd,
	) == false {
		return false
	}

	return true
}

func IsPartiallyWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
	vEnd time.Time,
) bool {
	if IsStartWithinTimeframe(
		timeframeStart, timeframeEnd, vStart,
	) == true {
		return true
	}
	if IsEndWithinTimeframe(
		timeframeStart, timeframeEnd, vEnd,
	) == true {
		return true
	}

	return false
}
