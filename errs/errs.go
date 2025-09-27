package errs

import (
	"errors"
	"fmt"
)

var (
	ErrKeyNotFound     error = errors.New("Key not found")
	ErrEndBeforeStart  error = errors.New("End is before start")
	ErrAlreadyRunning  error = errors.New("Tracker is already running")
	ErrNothingToEnd    error = errors.New("Nothing to end")
	ErrNothingToResume error = errors.New("Nothing to resume")
)

var ErrSIDNotFound error = errors.New("SID not found")

var ErrMissingProjectOrTaskSID error = errors.New(
	"'on' requires a projectID/taskID, " +
		"e.g. 'on myproject/mytask'",
)

var ErrMissingAttrOrVal error = errors.New(
	"'with' requires an attribute and a value, " +
		"e.g. 'with note \"Issue ID: 123\"'",
)

var ErrInvalidSID error = errors.New(
	"The Simplified-ID (SID) may only contain letters, numbers, dashes, " +
		"underscores, and periods",
)

var ErrNoteTooLarge error = errors.New(
	"The note is too large",
)

var ErrSIDTooLarge error = errors.New(
	"The SID is too large",
)

var ErrDisplayNameTooLarge error = errors.New(
	"The display name is too large",
)

var ErrInvalidColor error = errors.New(
	"The color must be in hex format (#FFFFFF)",
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
