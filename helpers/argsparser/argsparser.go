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
	endCommands    = []string{"end", "stop", "pause"}
)

type attribute struct {
	TakesValue bool
	Set        func(p *parser, value string) error
}

var attributes = map[string]attribute{
	"note": {
		TakesValue: true,
		Set: func(p *parser, value string) error {
			if p.pa.Note != "" {
				return errs.ErrRepeatedNote
			}

			p.pa.Note = value

			return nil
		},
	},
}

type ParsedArgs struct {
	ProjectSID     string `validate:"omitempty,sid,max=32"`
	TaskSID        string `validate:"omitempty,sid,max=32"`
	Note           string `validate:"max=65536"`
	TimestampStart string
	timestampStart time.Time
	TimestampEnd   string
	timestampEnd   time.Time
	processed      bool
}

type parser struct {
	pa        *ParsedArgs
	args      []string
	bareIsEnd bool
}

func isKeyword(word string) bool {
	return slices.Contains(projectWords, word) == true ||
		slices.Contains(attributeWords, word) == true ||
		slices.Contains(endWords, word) == true
}

func endsByDefault(command string) bool {
	return slices.Contains(endCommands, strings.ToLower(command))
}

func Parse(command string, args []string) (*ParsedArgs, error) {
	p := &parser{
		pa:        new(ParsedArgs),
		args:      args,
		bareIsEnd: endsByDefault(command),
	}

	for i := 0; i < len(args); {
		next, err := p.step(i)
		if err != nil {
			return nil, err
		}

		i = next
	}

	return p.pa, nil
}

func (p *parser) step(i int) (int, error) {
	word := strings.ToLower(p.args[i])

	switch {
	case slices.Contains(noiseWords, word) == true:
		return i + 1, nil
	case slices.Contains(projectWords, word) == true:
		return p.parseProjectTask(i)
	case slices.Contains(attributeWords, word) == true:
		return p.parseAttribute(i)
	default:
		return p.parseTimestamps(i)
	}
}

func (p *parser) parseProjectTask(i int) (int, error) {
	if i+1 >= len(p.args) {
		return 0, errs.ErrMissingProjectOrTaskSID
	}

	projectSID, taskSID, found := strings.Cut(strings.ToLower(p.args[i+1]), "/")
	if found == false || projectSID == "" || taskSID == "" {
		return 0, errs.ErrMissingProjectOrTaskSID
	}

	if p.pa.ProjectSID != "" || p.pa.TaskSID != "" {
		return 0, errs.ErrRepeatedProjectOrTask
	}

	p.pa.ProjectSID = projectSID
	p.pa.TaskSID = taskSID

	return i + 2, nil
}

func (p *parser) parseAttribute(i int) (int, error) {
	if i+1 >= len(p.args) {
		return 0, errs.ErrMissingAttrOrVal
	}

	attr, known := attributes[strings.ToLower(p.args[i+1])]
	if known == false {
		return 0, errs.ErrUnknownAttr
	}

	if attr.TakesValue == false {
		if err := attr.Set(p, ""); err != nil {
			return 0, err
		}

		return i + 2, nil
	}

	if i+2 >= len(p.args) {
		return 0, errs.ErrMissingAttrOrVal
	}

	if err := attr.Set(p, p.args[i+2]); err != nil {
		return 0, err
	}

	return i + 3, nil
}

func (p *parser) parseTimestamps(i int) (int, error) {
	stop := p.spanEndsAt(i)

	if err := p.setBareTimestamp(
		strings.Join(p.args[i:stop], " "),
	); err != nil {
		return 0, err
	}

	if stop >= len(p.args) ||
		slices.Contains(endWords, strings.ToLower(p.args[stop])) == false {
		return stop, nil
	}

	endStop := p.spanEndsAt(stop + 1)

	if err := p.setTimestampEnd(
		strings.Join(p.args[stop+1:endStop], " "),
	); err != nil {
		return 0, err
	}

	return endStop, nil
}

func (p *parser) spanEndsAt(from int) int {
	for i := from; i < len(p.args); i++ {
		if isKeyword(strings.ToLower(p.args[i])) == true {
			return i
		}
	}

	return len(p.args)
}

func (p *parser) setBareTimestamp(span string) error {
	if p.bareIsEnd == true {
		return p.setTimestampEnd(span)
	}

	return p.setTimestampStart(span)
}

func (p *parser) setTimestampStart(span string) error {
	if span == "" {
		return nil
	}

	if p.pa.TimestampStart != "" {
		return errs.ErrRepeatedTimestamp
	}

	p.pa.TimestampStart = span

	return nil
}

func (p *parser) setTimestampEnd(span string) error {
	if span == "" {
		return nil
	}

	if p.pa.TimestampEnd != "" {
		return errs.ErrRepeatedTimestamp
	}

	p.pa.TimestampEnd = span

	return nil
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
