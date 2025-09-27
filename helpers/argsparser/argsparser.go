package argsparser

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mrusme/zeit/errs"
	"github.com/mrusme/zeit/helpers/log"
	"github.com/mrusme/zeit/helpers/timestamp"
)

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
					return nil, errs.ErrMissingProjectOrTaskSID
				} else {
					i += 1
					continue
				}
			} else {
				return nil, errs.ErrMissingProjectOrTaskSID
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
				return nil, errs.ErrMissingAttrOrVal
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
		for _, err := range err.(validator.ValidationErrors) {
			if err.Tag() == "alphanum" {
				return errs.ErrSIDOnlyAlphanum
			} else if err.Field() == "Note" && err.Tag() == "max" {
				return errs.ErrNoteTooLarge
			}
		}

		return err
	}

	if pa.TimestampStart != "" {
		ts, err := timestamp.Parse(pa.TimestampStart)
		if err != nil {
			return &errs.ErrParsingTimestamp{
				Message:   err.Error(),
				Timestamp: pa.TimestampStart,
			}
		}

		pa.timestampStart = ts.Time

		if ts.IsRange == true {
			pa.timestampEnd = ts.ToTime
		}
	}

	if pa.TimestampEnd != "" && pa.timestampEnd.IsZero() {
		ts, err := timestamp.Parse(pa.TimestampEnd)
		if err != nil {
			return &errs.ErrParsingTimestamp{
				Message:   err.Error(),
				Timestamp: pa.TimestampEnd,
			}
		}

		pa.timestampEnd = ts.Time
	}

	if pa.timestampEnd.IsZero() == false &&
		pa.timestampEnd.Before(pa.timestampStart) {
		return &errs.ErrParsingTimestamp{
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

func POP(
	cmdName string,
	flags *ParsedArgs,
	args []string,
	logger *log.Logger,
) (*ParsedArgs, error) {
	var pargs *ParsedArgs
	var err error

	if pargs, err = Parse(cmdName, args); err != nil {
		return nil, err
	}

	pargs.OverrideWith(flags)

	if logger != nil {
		logger.Debug("Parsed args",
			"pargs", pargs,
			"GetTimestampStart", pargs.GetTimestampStart(),
			"GetTimestampEnd", pargs.GetTimestampEnd(),
		)
	}

	if err = pargs.Process(); err != nil {
		return nil, err
	}

	if logger != nil {
		logger.Debug("Processed args",
			"pargs", pargs,
			"GetTimestampStart", pargs.GetTimestampStart(),
			"GetTimestampEnd", pargs.GetTimestampEnd(),
		)
	}

	return pargs, nil
}
