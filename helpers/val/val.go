package val

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mrusme/zeit/errs"
)

func Validate(s interface{}) error {
	var err error

	validate := validator.New()
	validate.RegisterValidation("sid", IsValidSID)
	validate.RegisterValidation("timestamp_start", IsValidTimestampStart)
	validate.RegisterValidation("timestamp_end", IsValidTimestampEnd)
	if err = validate.Struct(s); err != nil {
		return TransformValidationError(err)
	}

	return nil
}

func IsValidSID(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if value == "edit" {
		// The project/task command has an `edit` subcommand, hence we prohibit a
		// project/task to be named "edit". While it would be totally doable to
		// use `zeit project Edit` to avoid calling the `edit` command, it is
		// way too much effort to explain this to the average user and it would only
		// lead to confusion and issue reports.
		return false
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`)

	return re.MatchString(value)
}

func IsValidTimestampStart(fl validator.FieldLevel) bool {
	fi := fl.Field().Interface()
	param := fl.Param()
	ofi := fl.Parent().FieldByName(param).Interface()

	ts, ok1 := fi.(time.Time)
	ots, ok2 := ofi.(time.Time)
	if ok1 == false || ok2 == false {
		return false
	}

	return ts.IsZero() == false &&
		(ts.Before(ots) || ots.IsZero() == true)
}

func IsValidTimestampEnd(fl validator.FieldLevel) bool {
	fi := fl.Field().Interface()
	param := fl.Param()
	ofi := fl.Parent().FieldByName(param).Interface()

	ts, ok1 := fi.(time.Time)
	ots, ok2 := ofi.(time.Time)
	if ok1 == false || ok2 == false {
		return false
	}

	return ts.IsZero() == true ||
		(ts.After(ots) && ots.IsZero() == false)
}

func TransformValidationError(err error) error {
	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "timestamp_start":
			return errs.ErrInvalidTimestampStart
		case "timestamp_end":
			return errs.ErrInvalidTimestampEnd
		case "required":
			switch err.Field() {
			case "ProjectSID":
				return errs.ErrProjectSIDRequired
			case "TaskSID":
				return errs.ErrTaskSIDRequired
			}
		case "sid":
			return errs.ErrInvalidSID
		case "max":
			switch err.Field() {
			case "Note":
				return errs.ErrNoteTooLarge
			case "ProjectSID", "TaskSID":
				return errs.ErrSIDTooLarge
			case "DisplayName":
				return errs.ErrDisplayNameTooLarge
			}
		case "hexcolor":
			return errs.ErrInvalidColor
		}
	}

	return err
}
