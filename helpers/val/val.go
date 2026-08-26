package val

import (
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"xn--gckvb8fzb.com/zeit/errs"
)

const (
	VALID_SID_REGEX     = `^[a-zA-Z0-9\-\_\.]+$`
	VALID_SID_REGEX_NEG = `[^a-zA-Z0-9\-\_\.]`
)

var (
	validSIDRegex        = regexp.MustCompile(VALID_SID_REGEX)
	invalidSIDCharsRegex = regexp.MustCompile(VALID_SID_REGEX_NEG)
	underscoreRunRegex   = regexp.MustCompile(`_+`)
)

var getValidator = sync.OnceValues(func() (*validator.Validate, error) {
	validate := validator.New()

	if err := validate.RegisterValidation("sid", IsValidSID); err != nil {
		return nil, err
	}
	if err := validate.RegisterValidation(
		"timestamp_start", IsValidTimestampStart,
	); err != nil {
		return nil, err
	}
	if err := validate.RegisterValidation(
		"timestamp_end", IsValidTimestampEnd,
	); err != nil {
		return nil, err
	}

	return validate, nil
})

func Validate(s interface{}) error {
	validate, err := getValidator()
	if err != nil {
		return err
	}

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

	return validSIDRegex.MatchString(value)
}

func ConvertTextToSID(txt string) string {
	tmp := strings.ToLower(invalidSIDCharsRegex.ReplaceAllString(txt, "_"))

	tmp = underscoreRunRegex.ReplaceAllString(tmp, "_")

	if len(tmp) > 32 {
		tmp = tmp[:32]
	}

	tmp = strings.Trim(tmp, "_")

	return tmp
}

func FitDisplayName(dn string) string {
	if runes := []rune(dn); len(runes) > 32 {
		return string(runes[:32])
	}
	return dn
}

func ConvertSIDToDisplayName(sid string) string {
	runes := []rune(sid)
	if len(runes) == 0 {
		return sid
	}

	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
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
	verrs, ok := err.(validator.ValidationErrors)
	if ok == false {
		return err
	}

	for _, err := range verrs {
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
