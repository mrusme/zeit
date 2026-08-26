package argsparser

import (
	"slices"
	"strings"
	"time"

	"xn--gckvb8fzb.com/zeit/errs"
	"xn--gckvb8fzb.com/zeit/helpers/log"
	"xn--gckvb8fzb.com/zeit/helpers/timestamp"
	"xn--gckvb8fzb.com/zeit/helpers/val"
)

var (
	noiseWords     = []string{"block", "working", "work", "wrk", "all", "at", "from"}
	projectWords   = []string{"on", "to", "of"}
	attributeWords = []string{"with", "w"}
	endWords       = []string{"end", "ends", "ended", "til", "until"}
)

type ParsedArgs struct {
	ProjectSID     string    `validate:"omitempty,required_with=TaskSID,sid,max=32"`
	TaskSID        string    `validate:"omitempty,required_with=ProjectSID,sid,max=32"`
	Note           string    `validate:"max=65536"`
	TimestampStart string    `validate:""`
	timestampStart time.Time `validate:""`
	TimestampEnd   string    `validate:""`
	timestampEnd   time.Time `validate:""`
	processed      bool
}

func isKeyword(word string) bool {
	return slices.Contains(projectWords, word) == true ||
		slices.Contains(attributeWords, word) == true ||
		slices.Contains(endWords, word) == true
}

func timestampEndsAt(args []string, from int) int {
	for j := from; j < len(args); j++ {
		if isKeyword(strings.ToLower(args[j])) == true {
			return j
		}
	}

	return len(args)
}

func Parse(command string, args []string) (*ParsedArgs, error) {
	pa := new(ParsedArgs)

	for i := 0; i < len(args); i++ {
		word := strings.ToLower(args[i])
		if slices.Contains(noiseWords, word) == true {
			continue
		} else if slices.Contains(projectWords, word) == true {
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
		} else if slices.Contains(attributeWords, word) == true {
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
			stop := timestampEndsAt(args, i)
			pa.TimestampStart = strings.Join(args[i:stop], " ")

			if stop < len(args) &&
				slices.Contains(endWords, strings.ToLower(args[stop])) == true {
				endStop := timestampEndsAt(args, stop+1)
				pa.TimestampEnd = strings.Join(args[stop+1:endStop], " ")
				stop = endStop
			}

			i = stop - 1
			continue
		}
	}

	return pa, nil
}

func (pa *ParsedArgs) Process() error {
	var err error

	if err = val.Validate(*pa); err != nil {
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
