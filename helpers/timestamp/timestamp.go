package timestamp

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"xn--gckvb8fzb.com/zeit/errs"
)

type Timestamp struct {
	Time    time.Time
	ToTime  time.Time
	IsRange bool
}

var periodRegex = regexp.MustCompile(
	`(?m)^(?:(this|current|last|previous)\s+)?(hour|day|week|month|quarter|year|decade|century)$`,
)

type Period struct {
	start func(t time.Time) time.Time
	step  func(t time.Time, n int) time.Time
	key   func(t time.Time) string
}

func (p Period) Start(t time.Time) time.Time {
	return p.start(t)
}

func (p Period) Next(t time.Time) time.Time {
	return p.step(t, 1)
}

func (p Period) Previous(t time.Time) time.Time {
	return p.step(t, -1)
}

func (p Period) Key(t time.Time) string {
	return p.key(t)
}

var periods = map[string]Period{
	"hour": {
		start: func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(),
				0, 0, 0, t.Location())
		},
		step: func(t time.Time, n int) time.Time {
			return t.Add(time.Duration(n) * time.Hour)
		},
		key: func(t time.Time) string {
			return t.Format("2006-01-02T15")
		},
	},
	"day": {
		start: func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		},
		step: func(t time.Time, n int) time.Time {
			return t.AddDate(0, 0, n)
		},
		key: func(t time.Time) string {
			return t.Format(time.DateOnly)
		},
	},
	"week": {
		start: func(t time.Time) time.Time {
			daysToMonday := int(t.Weekday()-time.Monday+7) % 7

			return time.Date(t.Year(), t.Month(), t.Day()-daysToMonday,
				0, 0, 0, 0, t.Location())
		},
		step: func(t time.Time, n int) time.Time {
			return t.AddDate(0, 0, 7*n)
		},
		key: func(t time.Time) string {
			year, week := t.ISOWeek()

			return fmt.Sprintf("%d-W%02d", year, week)
		},
	},
	"month": {
		start: func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		},
		step: func(t time.Time, n int) time.Time {
			return t.AddDate(0, n, 0)
		},
		key: func(t time.Time) string {
			return t.Format("2006-01")
		},
	},
	"quarter": {
		start: quarterStart,
		step: func(t time.Time, n int) time.Time {
			return t.AddDate(0, 3*n, 0)
		},
		key: func(t time.Time) string {
			return fmt.Sprintf("%d-Q%d", t.Year(), quarterOf(t))
		},
	},
	"year": {
		start: func(t time.Time) time.Time {
			return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
		},
		step: func(t time.Time, n int) time.Time {
			return t.AddDate(n, 0, 0)
		},
		key: func(t time.Time) string {
			return t.Format("2006")
		},
	},
}

func GetPeriod(name string) (Period, error) {
	period, ok := periods[strings.ToLower(name)]
	if ok == false {
		return Period{}, errs.ErrNotATimeframe
	}

	return period, nil
}

func quarterOf(t time.Time) int {
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
	return (int(t.Month())-1)/3 + 1
}

func quarterStart(t time.Time) time.Time {
	month := (quarterOf(t)-1)*3 + 1

	return time.Date(t.Year(), time.Month(month), 1, 0, 0, 0, 0, t.Location())
}

func ParsePeriod(str string) (*Timestamp, error) {
	matches := periodRegex.FindStringSubmatch(str)
	if len(matches) != 3 {
		return nil, errors.New("No period found")
	}

	frame := strings.ToLower(matches[1])
	name := strings.ToLower(matches[2])

	ts := new(Timestamp)
	ts.IsRange = true

	period, ok := periods[name]
	if ok == false {
		return ts, nil
	}

	start := period.Start(time.Now())
	if frame == "last" || frame == "previous" {
		start = period.Previous(start)
	}

	ts.Time = start
	ts.ToTime = period.Next(start).Add(-time.Second)

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

func ClipToTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
	vEnd time.Time,
) (time.Time, time.Time) {
	start := vStart
	if timeframeStart.IsZero() == false && start.Before(timeframeStart) == true {
		start = timeframeStart
	}

	end := vEnd
	if timeframeEnd.IsZero() == false && end.After(timeframeEnd) == true {
		end = timeframeEnd
	}

	return start, end
}

func DurationWithinTimeframe(
	timeframeStart time.Time,
	timeframeEnd time.Time,
	vStart time.Time,
	vEnd time.Time,
) time.Duration {
	start, end := ClipToTimeframe(timeframeStart, timeframeEnd, vStart, vEnd)

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
