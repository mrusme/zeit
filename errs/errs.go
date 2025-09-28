package errs

import (
	"errors"
	"fmt"
)

var (
	ErrKeyNotFound error = errors.New(
		"Key not found",
	)

	ErrAlreadyRunning error = errors.New(
		"Tracker is already running",
	)

	ErrNothingToEnd error = errors.New(
		"Nothing to end",
	)

	ErrNothingToResume error = errors.New(
		"Nothing to resume",
	)

	ErrSIDNotFound error = errors.New(
		"SID not found",
	)

	ErrMissingProjectOrTaskSID error = errors.New(
		"'on' requires a projectID/taskID, " +
			"e.g. 'on myproject/mytask'",
	)

	ErrMissingAttrOrVal error = errors.New(
		"'with' requires an attribute and a value, " +
			"e.g. 'with note \"Issue ID: 123\"'",
	)

	ErrInvalidSID error = errors.New(
		"The Simplified-ID (SID) may only contain letters, numbers, dashes, " +
			"underscores, and periods. Certain reserved keywords like 'edit' are " +
			"not allowed.",
	)

	ErrNoteTooLarge error = errors.New(
		"The note is too large",
	)

	ErrSIDTooLarge error = errors.New(
		"The SID is too large",
	)

	ErrProjectSIDRequired error = errors.New(
		"A project SID is required",
	)

	ErrTaskSIDRequired error = errors.New(
		"A task SID is required",
	)

	ErrDisplayNameTooLarge error = errors.New(
		"The display name is too large",
	)

	ErrInvalidColor error = errors.New(
		"The color must be in hex format (e.g. #FFFFFF)",
	)

	ErrInvalidTimestampStart error = errors.New(
		"The start time/date must be before the end time/date and not be empty",
	)

	ErrInvalidTimestampEnd error = errors.New(
		"The end time/date must be after the start time/date",
	)
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
