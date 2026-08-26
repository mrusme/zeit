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
	`(?m)^(?:(this|current|last|previous)\s+)?(hour|day|week|month|quarter|year|decade|century)$`,
)

func ParsePeriod(str string) (*Timestamp, error) {
	var frame string
	var period string
	var now time.Time = time.Now()

	ts := new(Timestamp)

	matches := periodRegex.FindStringSubmatch(str)

	if len(matches) != 3 {
		return nil, errors.New("No period found")
	}

	frame = strings.ToLower(matches[1])
	period = strings.ToLower(matches[2])

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
		ts.ToTime = ts.Time.AddDate(0, 0, 1).Add(-time.Second)
	case "week":
		daysToMonday := int(now.Weekday()-time.Monday+7) % 7
		ts.Time = time.Date(now.Year(), now.Month(), now.Day()-daysToMonday,
			0, 0, 0, 0, now.Location())
		if previousPeriod {
			ts.Time = ts.Time.AddDate(0, 0, -7)
		}
		ts.ToTime = ts.Time.AddDate(0, 0, 7).Add(-time.Second)
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

	qStart := time.Date(now.Year(), time.Month(quarterStartMonth), 1,
		0, 0, 0, 0, now.Location())
	if last {
		qStart = qStart.AddDate(0, -3, 0)
	}
	qEnd := qStart.AddDate(0, 3, 0).Add(-time.Second)

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

func IsWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	v time.Time,
) bool {
	if timeframeStart.IsZero() == false && v.Before(timeframeStart) {
		return false
	}

	if timeframeEnd.IsZero() == false && v.After(timeframeEnd) {
		return false
	}

	return true
}

func IsStartWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
) bool {
	return IsWithinTimeframe(timeframeStart, timeframeEnd, vStart)
}

func endOrNow(vEnd time.Time) time.Time {
	if vEnd.IsZero() == true {
		return time.Now()
	}

	return vEnd
}

func IsEndWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vEnd time.Time,
) bool {
	return IsWithinTimeframe(timeframeStart, timeframeEnd, endOrNow(vEnd))
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

func DurationWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
	vEnd time.Time,
) time.Duration {
	start := vStart
	if timeframeStart.IsZero() == false && start.Before(timeframeStart) == true {
		start = timeframeStart
	}

	end := vEnd
	if timeframeEnd.IsZero() == false && end.After(timeframeEnd) == true {
		end = timeframeEnd
	}

	if end.After(start) == false {
		return 0
	}

	return end.Sub(start)
}

func IsPartiallyWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
	vEnd time.Time,
) bool {
	if timeframeStart.IsZero() == false &&
		endOrNow(vEnd).Before(timeframeStart) == true {
		return false
	}

	if timeframeEnd.IsZero() == false && vStart.After(timeframeEnd) == true {
		return false
	}

	return true
}
