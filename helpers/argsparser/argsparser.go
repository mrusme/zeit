package argsparser

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mrusme/zeit/helpers/timestamp"
)

var ErrMissingProjectOrTaskID error = errors.New(
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
	ProjectID      string
	TaskID         string
	Note           string
	TimestampStart time.Time
	TimestampEnd   time.Time
}

func Parse(command string, args []string) (*ParsedArgs, error) {
	var tstampStart, tstampEnd string
	var err error
	pa := new(ParsedArgs)

	for i := 0; i < len(args); i++ {
		word := strings.ToLower(args[i])
		if word == "block" || word == "working" || word == "work" || word == "wrk" {
			continue
		} else if word == "on" {
			if len(args) > i+1 {
				pst := strings.ToLower(args[i+1])
				found := false
				pa.ProjectID, pa.TaskID, found = strings.Cut(pst, "/")
				if found == false {
					return nil, ErrMissingProjectOrTaskID
				} else {
					i += 1
					continue
				}
			} else {
				return nil, ErrMissingProjectOrTaskID
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
			if word == "at" {
				continue
			}

			endMarker := -1
			for j := i; j < len(args); j++ {
				nextWord := strings.ToLower(args[j])
				endMarker = -1
				if nextWord == "end" || nextWord == "ends" || nextWord == "ended" {
					endMarker = j
					break
				}
			}

			if endMarker > -1 {
				tstampStart = strings.Join(args[i:endMarker], " ")
				tstampEnd = strings.Join(args[endMarker+1:], " ")
			} else {
				tstampStart = strings.Join(args[i:], " ")
			}
			break
		}
	}

	fmt.Printf("Project ID: %s\nTask ID: %s\nStart Timestamp: %s\nEnd Timestamp: %s\n",
		pa.ProjectID, pa.TaskID, tstampStart, tstampEnd)

	if tstampStart != "" {
		pa.TimestampStart, err = timestamp.Parse(tstampStart)
		if err != nil {
			return nil, &ErrParsingTimestamp{
				Message:   err.Error(),
				Timestamp: tstampStart,
			}
		}
	}

	if tstampEnd != "" {
		pa.TimestampEnd, err = timestamp.Parse(tstampEnd)
		if err != nil {
			return nil, &ErrParsingTimestamp{
				Message:   err.Error(),
				Timestamp: tstampEnd,
			}
		}
	}

	return pa, nil
}
