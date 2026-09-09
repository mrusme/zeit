package val

import (
	"errors"
	"strings"
	"testing"
	"time"

	"xn--gckvb8fzb.com/zeit/errs"
)

type sidHolder struct {
	ProjectSID string `validate:"required,sid,max=32"`
}

type noteHolder struct {
	Note string `validate:"max=8"`
}

type displayNameHolder struct {
	DisplayName string `validate:"max=4"`
}

type projectSIDHolder struct {
	ProjectSID string `validate:"required,sid,max=4"`
}

type taskSIDHolder struct {
	TaskSID string `validate:"required"`
}

type colorHolder struct {
	Color string `validate:"hexcolor"`
}

type timestampHolder struct {
	TimestampStart time.Time `validate:"timestamp_start=TimestampEnd"`
	TimestampEnd   time.Time `validate:"timestamp_end=TimestampStart"`
}

func TestValidateSID(t *testing.T) {
	tests := []struct {
		name string
		sid  string
		want error
	}{
		{name: "lowercase", sid: "myproject", want: nil},
		{name: "mixed case", sid: "MyProject", want: nil},
		{name: "digits", sid: "project2026", want: nil},
		{name: "dash underscore period", sid: "my-project_v1.2", want: nil},
		{name: "single character", sid: "a", want: nil},
		{name: "exactly 32 characters", sid: strings.Repeat("a", 32), want: nil},
		{name: "empty", sid: "", want: errs.ErrProjectSIDRequired},
		{name: "space", sid: "my project", want: errs.ErrInvalidSID},
		{name: "slash", sid: "my/project", want: errs.ErrInvalidSID},
		{name: "exclamation mark", sid: "project!", want: errs.ErrInvalidSID},
		{name: "non ascii", sid: "prüjekt", want: errs.ErrInvalidSID},
		{name: "reserved keyword edit", sid: "edit", want: errs.ErrInvalidSID},
		{name: "33 characters", sid: strings.Repeat("a", 33), want: errs.ErrSIDTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(sidHolder{ProjectSID: test.sid})
			if test.want == nil {
				if err != nil {
					t.Errorf("Validate(%q) = %v, want nil", test.sid, err)
				}
				return
			}
			if errors.Is(err, test.want) == false {
				t.Errorf("Validate(%q) = %v, want %v", test.sid, err, test.want)
			}
		})
	}
}

func TestValidateSIDErrorMapping(t *testing.T) {
	if err := Validate(projectSIDHolder{ProjectSID: ""}); errors.Is(
		err, errs.ErrProjectSIDRequired,
	) == false {
		t.Errorf("empty ProjectSID = %v, want ErrProjectSIDRequired", err)
	}

	if err := Validate(taskSIDHolder{TaskSID: ""}); errors.Is(
		err, errs.ErrTaskSIDRequired,
	) == false {
		t.Errorf("empty TaskSID = %v, want ErrTaskSIDRequired", err)
	}

	if err := Validate(projectSIDHolder{ProjectSID: "toolong"}); errors.Is(
		err, errs.ErrSIDTooLarge,
	) == false {
		t.Errorf("oversized ProjectSID = %v, want ErrSIDTooLarge", err)
	}

	if err := Validate(noteHolder{Note: "far too long"}); errors.Is(
		err, errs.ErrNoteTooLarge,
	) == false {
		t.Errorf("oversized Note = %v, want ErrNoteTooLarge", err)
	}

	if err := Validate(displayNameHolder{DisplayName: "toolong"}); errors.Is(
		err, errs.ErrDisplayNameTooLarge,
	) == false {
		t.Errorf("oversized DisplayName = %v, want ErrDisplayNameTooLarge", err)
	}

	if err := Validate(colorHolder{Color: "not-a-color"}); errors.Is(
		err, errs.ErrInvalidColor,
	) == false {
		t.Errorf("invalid Color = %v, want ErrInvalidColor", err)
	}

	if err := Validate(colorHolder{Color: "#AABBCC"}); err != nil {
		t.Errorf("valid Color = %v, want nil", err)
	}
}

func TestValidateTimestamps(t *testing.T) {
	early := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)
	late := time.Date(2026, time.August, 26, 17, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  error
	}{
		{name: "start before end", start: early, end: late, want: nil},
		{name: "open ended", start: early, end: time.Time{}, want: nil},
		{
			name:  "missing start",
			start: time.Time{},
			end:   late,
			want:  errs.ErrInvalidTimestampStart,
		},
		{
			name:  "missing both",
			start: time.Time{},
			end:   time.Time{},
			want:  errs.ErrInvalidTimestampStart,
		},
		{
			name:  "end before start",
			start: late,
			end:   early,
			want:  errs.ErrInvalidTimestampStart,
		},
		{
			name:  "end equals start",
			start: early,
			end:   early,
			want:  errs.ErrInvalidTimestampStart,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(timestampHolder{
				TimestampStart: test.start,
				TimestampEnd:   test.end,
			})
			if test.want == nil {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if errors.Is(err, test.want) == false {
				t.Errorf("Validate() = %v, want %v", err, test.want)
			}
		})
	}
}

type modelSIDHolder struct {
	SID string `validate:"required,sid,max=32"`
}

func TestValidateModelSID(t *testing.T) {
	tests := []struct {
		name string
		sid  string
		want error
	}{
		{name: "valid", sid: "myproject", want: nil},
		{name: "empty", sid: "", want: errs.ErrSIDRequired},
		{name: "invalid", sid: "my project", want: errs.ErrInvalidSID},
		{name: "oversized", sid: strings.Repeat("a", 33), want: errs.ErrSIDTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(modelSIDHolder{SID: test.sid})

			if test.want == nil {
				if err != nil {
					t.Errorf("Validate(%q) = %v, want nil", test.sid, err)
				}
				return
			}

			if errors.Is(err, test.want) == false {
				t.Errorf("Validate(%q) = %v, want %v", test.sid, err, test.want)
			}
		})
	}
}

func TestValidateNonStruct(t *testing.T) {
	if err := Validate(42); err == nil {
		t.Errorf("Validate(42) = nil, want an error")
	}
}

func TestTransformValidationErrorPassesThroughUnknownErrors(t *testing.T) {
	sentinel := errors.New("not a validation error")

	if got := TransformValidationError(sentinel); errors.Is(got, sentinel) == false {
		t.Errorf("TransformValidationError() = %v, want %v", got, sentinel)
	}
}

func TestConvertTextToSID(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "already a sid", text: "myproject", want: "myproject"},
		{name: "lowercases", text: "MyProject", want: "myproject"},
		{name: "spaces become underscores", text: "Old Project", want: "old_project"},
		{name: "collapses runs", text: "a!!!b", want: "a_b"},
		{name: "trims leading and trailing", text: " padded ", want: "padded"},
		{name: "keeps dashes and periods", text: "v1.2-beta", want: "v1.2-beta"},
		{name: "non ascii", text: "日本語", want: ""},
		{name: "empty", text: "", want: ""},
		{
			name: "truncates to 32",
			text: strings.Repeat("a", 40),
			want: strings.Repeat("a", 32),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertTextToSID(test.text)
			if got != test.want {
				t.Errorf("ConvertTextToSID(%q) = %q, want %q", test.text, got, test.want)
			}
		})
	}
}

func TestConvertTextToSIDProducesValidSIDs(t *testing.T) {
	texts := []string{
		"Old Project",
		"a!!!b",
		"v1.2-beta",
		strings.Repeat("a", 40),
		"Ünïcödé Näme",
	}

	for _, text := range texts {
		sid := ConvertTextToSID(text)
		if sid == "" {
			continue
		}

		if err := Validate(sidHolder{ProjectSID: sid}); err != nil {
			t.Errorf("ConvertTextToSID(%q) = %q, which fails validation: %v",
				text, sid, err)
		}
	}
}

func TestFitDisplayName(t *testing.T) {
	tests := []struct {
		name string
		dn   string
		want string
	}{
		{name: "short", dn: "Short", want: "Short"},
		{name: "empty", dn: "", want: ""},
		{
			name: "exactly 32 ascii",
			dn:   strings.Repeat("a", 32),
			want: strings.Repeat("a", 32),
		},
		{
			name: "33 ascii",
			dn:   strings.Repeat("a", 33),
			want: strings.Repeat("a", 32),
		},
		{
			name: "multibyte is cut by rune",
			dn:   strings.Repeat("日", 40),
			want: strings.Repeat("日", 32),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := FitDisplayName(test.dn)
			if got != test.want {
				t.Errorf("FitDisplayName(%q) = %q, want %q", test.dn, got, test.want)
			}
		})
	}
}

func TestFitDisplayNameKeepsValidUTF8(t *testing.T) {
	for prefix := range 8 {
		dn := strings.Repeat("x", prefix) + strings.Repeat("日", 40)

		got := FitDisplayName(dn)
		if utf8Valid(got) == false {
			t.Errorf("FitDisplayName(%q) produced invalid UTF-8: %q", dn, got)
		}
	}
}

func TestConvertSIDToDisplayName(t *testing.T) {
	tests := []struct {
		name string
		sid  string
		want string
	}{
		{name: "lowercase", sid: "myproject", want: "Myproject"},
		{name: "already capitalised", sid: "Myproject", want: "Myproject"},
		{name: "underscores are kept", sid: "my_project", want: "My_project"},
		{name: "leading digit", sid: "2026-review", want: "2026-review"},
		{name: "multibyte", sid: "über", want: "Über"},
		{name: "single character", sid: "a", want: "A"},
		{name: "empty", sid: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertSIDToDisplayName(test.sid)
			if got != test.want {
				t.Errorf("ConvertSIDToDisplayName(%q) = %q, want %q",
					test.sid, got, test.want)
			}
		})
	}
}

func utf8Valid(s string) bool {
	for _, r := range s {
		if r == '�' {
			return false
		}
	}
	return true
}
