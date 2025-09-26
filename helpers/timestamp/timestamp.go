package timestamp

import (
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
)

type Timestamp struct {
	Time    time.Time
	ToTime  time.Time
	IsRange bool
}

func Parse(str string) (*Timestamp, error) {
	var err error
	var dt date.Date

	ts := new(Timestamp)

	cfg := dateparser.Configuration{
		DefaultTimezone: time.Local,
	}

	if dt, err = dateparser.Parse(&cfg, str); err != nil {
		return nil, err
	}

	ts.Time = dt.Time

	return ts, nil
}
