package val

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/mrusme/zeit/errs"
)

func Validate(s interface{}) error {
	var err error

	validate := validator.New()
	validate.RegisterValidation("sid", IsValidSID)
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

func TransformValidationError(err error) error {
	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
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
