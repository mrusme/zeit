package timestamp

import (
	"time"

	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
)

func Parse(str string) (time.Time, error) {
	var err error
	var dt date.Date

	cfg := dateparser.Configuration{
		DefaultTimezone: time.Local,
	}

	if dt, err = dateparser.Parse(&cfg, str); err != nil {
		return time.Now(), err
	}

	return dt.Time, nil
}
