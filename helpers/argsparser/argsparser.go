package argsparser

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mrusme/zeit/helpers/timestamp"
)

var ErrMissingProjectOrTaskSID error = errors.New(
	"'on' requires a projectID/taskID, " +
		"e.g. 'on myproject/mytask'",
)

var ErrMissingAttrOrVal error = errors.New(
	"'with' requires an attribute and a value, " +
		"e.g. 'with note \"Issue ID: 123\"'",
)

type ErrParsingTimestamp struct {
	Message   string
	Timestamp string
}

func (e *ErrParsingTimestamp) Error() string {
	return fmt.Sprintf(
		"Error parsing timestamp: %s\nTimestamp: %s\n",
		e.Message,
		e.Timestamp,
	)
}

type ParsedArgs struct {
	ProjectSID     string    `validate:"omitempty,required_with=TaskSID,alphanum,max=64"`
	TaskSID        string    `validate:"omitempty,required_with=ProjectSID,alphanum,max=64"`
	Note           string    `validate:"max=65536"`
	TimestampStart string    `validate:""`
	timestampStart time.Time `validate:""`
	TimestampEnd   string    `validate:""`
	timestampEnd   time.Time `validate:""`
	processed      bool
}

func Parse(command string, args []string) (*ParsedArgs, error) {
	pa := new(ParsedArgs)

	for i := 0; i < len(args); i++ {
		word := strings.ToLower(args[i])
		if word == "block" ||
			word == "working" || word == "work" || word == "wrk" ||
			word == "all" {
			continue
		} else if word == "on" || word == "to" || word == "of" {
			if len(args) > i+1 {
				pst := strings.ToLower(args[i+1])
				found := false
				pa.ProjectSID, pa.TaskSID, found = strings.Cut(pst, "/")
				if found == false {
					return nil, ErrMissingProjectOrTaskSID
				} else {
					i += 1
					continue
				}
			} else {
				return nil, ErrMissingProjectOrTaskSID
			}
		} else if word == "with" || word == "w" {
			if len(args) > i+2 {
				attr := strings.ToLower(args[i+1])
				val := args[i+2]

				switch attr {
				case "note":
					pa.Note = val
				}

				i += 2
				continue
			} else {
				return nil, ErrMissingAttrOrVal
			}
		} else {
			if word == "at" || word == "from" {
				continue
			}

			endMarker := -1
			for j := i; j < len(args); j++ {
				nextWord := strings.ToLower(args[j])
				endMarker = -1
				if nextWord == "end" || nextWord == "ends" || nextWord == "ended" ||
					nextWord == "til" || nextWord == "until" {
					endMarker = j
					break
				}
			}

			if endMarker > -1 {
				pa.TimestampStart = strings.Join(args[i:endMarker], " ")
				pa.TimestampEnd = strings.Join(args[endMarker+1:], " ")
			} else {
				pa.TimestampStart = strings.Join(args[i:], " ")
			}
			break
		}
	}

	return pa, nil
}

func (pa *ParsedArgs) Process() error {
	var err error

	validate := validator.New()
	if err = validate.Struct(*pa); err != nil {
		return err
	}

	if pa.TimestampStart != "" {
		ts, err := timestamp.Parse(pa.TimestampStart)
		if err != nil {
			return &ErrParsingTimestamp{
				Message:   err.Error(),
				Timestamp: pa.TimestampStart,
			}
		}

		pa.timestampStart = ts.Time
	}

	if pa.TimestampEnd != "" {
		ts, err := timestamp.Parse(pa.TimestampEnd)
		if err != nil {
			return &ErrParsingTimestamp{
				Message:   err.Error(),
				Timestamp: pa.TimestampEnd,
			}
		}

		pa.timestampEnd = ts.Time
	}

	if pa.timestampEnd.IsZero() == false &&
		pa.timestampEnd.Before(pa.timestampStart) {
		return &ErrParsingTimestamp{
			Message:   "End is before start",
			Timestamp: pa.TimestampEnd,
		}
	}

	pa.processed = true
	return nil
}

func (pa *ParsedArgs) WasProcessed() bool {
	return pa.processed
}

func (pa *ParsedArgs) GetTimestampStart() time.Time {
	return pa.timestampStart
}

func (pa *ParsedArgs) GetTimestampEnd() time.Time {
	return pa.timestampEnd
}

func (pa *ParsedArgs) OverrideWith(spa *ParsedArgs) {
	// TODO: Maybe use https://github.com/darccio/mergo ?

	if spa.ProjectSID != "" {
		pa.ProjectSID = spa.ProjectSID
	}
	if spa.TaskSID != "" {
		pa.TaskSID = spa.TaskSID
	}
	if spa.Note != "" {
		pa.Note = spa.Note
	}
	if spa.TimestampStart != "" {
		pa.TimestampStart = spa.TimestampStart
	}
	if spa.TimestampEnd != "" {
		pa.TimestampEnd = spa.TimestampEnd
	}

	return
}
