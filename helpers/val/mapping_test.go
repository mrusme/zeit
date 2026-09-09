package val_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"xn--gckvb8fzb.com/zeit/errs"
	"xn--gckvb8fzb.com/zeit/helpers/val"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/models/project"
	"xn--gckvb8fzb.com/zeit/models/task"
)

var friendlyErrors = []error{
	errs.ErrSIDRequired,
	errs.ErrProjectSIDRequired,
	errs.ErrTaskSIDRequired,
	errs.ErrInvalidSID,
	errs.ErrSIDTooLarge,
	errs.ErrNoteTooLarge,
	errs.ErrDisplayNameTooLarge,
	errs.ErrInvalidColor,
	errs.ErrInvalidTimestampStart,
	errs.ErrInvalidTimestampEnd,
}

func isFriendly(err error) bool {
	for _, friendly := range friendlyErrors {
		if errors.Is(err, friendly) == true {
			return true
		}
	}

	return false
}

func newProject(t *testing.T) *project.Project {
	t.Helper()

	pj, err := project.New("owner", "myproject")
	if err != nil {
		t.Fatalf("project.New() = %s", err)
	}

	return pj
}

func newTask(t *testing.T) *task.Task {
	t.Helper()

	tk, err := task.New("owner", "myproject", "mytask")
	if err != nil {
		t.Fatalf("task.New() = %s", err)
	}

	return tk
}

func newBlock(t *testing.T) *block.Block {
	t.Helper()

	b, err := block.New("owner")
	if err != nil {
		t.Fatalf("block.New() = %s", err)
	}

	b.ProjectSID = "myproject"
	b.TaskSID = "mytask"
	b.TimestampStart = time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)
	b.TimestampEnd = b.TimestampStart.Add(time.Hour)

	return b
}

func TestEveryModelViolationMapsToAFriendlyError(t *testing.T) {
	oversized := strings.Repeat("a", 33)

	tests := []struct {
		name  string
		build func(t *testing.T) any
	}{
		{
			name: "project without a sid",
			build: func(t *testing.T) any {
				pj := newProject(t)
				pj.SID = ""
				return *pj
			},
		},
		{
			name: "project with an invalid sid",
			build: func(t *testing.T) any {
				pj := newProject(t)
				pj.SID = "my project"
				return *pj
			},
		},
		{
			name: "project with an oversized sid",
			build: func(t *testing.T) any {
				pj := newProject(t)
				pj.SID = oversized
				return *pj
			},
		},
		{
			name: "project with an oversized display name",
			build: func(t *testing.T) any {
				pj := newProject(t)
				pj.DisplayName = oversized
				return *pj
			},
		},
		{
			name: "project without a color",
			build: func(t *testing.T) any {
				pj := newProject(t)
				pj.Color = ""
				return *pj
			},
		},
		{
			name: "project with an invalid color",
			build: func(t *testing.T) any {
				pj := newProject(t)
				pj.Color = "red"
				return *pj
			},
		},
		{
			name: "task without a sid",
			build: func(t *testing.T) any {
				tk := newTask(t)
				tk.SID = ""
				return *tk
			},
		},
		{
			name: "task with an oversized sid",
			build: func(t *testing.T) any {
				tk := newTask(t)
				tk.SID = oversized
				return *tk
			},
		},
		{
			name: "task without a project sid",
			build: func(t *testing.T) any {
				tk := newTask(t)
				tk.ProjectSID = ""
				return *tk
			},
		},
		{
			name: "task with an invalid project sid",
			build: func(t *testing.T) any {
				tk := newTask(t)
				tk.ProjectSID = "my project"
				return *tk
			},
		},
		{
			name: "task with an oversized display name",
			build: func(t *testing.T) any {
				tk := newTask(t)
				tk.DisplayName = oversized
				return *tk
			},
		},
		{
			name: "task with an invalid color",
			build: func(t *testing.T) any {
				tk := newTask(t)
				tk.Color = "red"
				return *tk
			},
		},
		{
			name: "block without a project sid",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.ProjectSID = ""
				return *b
			},
		},
		{
			name: "block without a task sid",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.TaskSID = ""
				return *b
			},
		},
		{
			name: "block with an invalid task sid",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.TaskSID = "my task"
				return *b
			},
		},
		{
			name: "block with an oversized project sid",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.ProjectSID = oversized
				return *b
			},
		},
		{
			name: "block with an oversized note",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.Note = strings.Repeat("a", 65537)
				return *b
			},
		},
		{
			name: "block without a start",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.TimestampStart = time.Time{}
				return *b
			},
		},
		{
			name: "block ending before it starts",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.TimestampEnd = b.TimestampStart.Add(-time.Hour)
				return *b
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := val.Validate(test.build(t))
			if err == nil {
				t.Fatalf("Validate() = nil, want an error")
			}

			if isFriendly(err) == false {
				t.Errorf("Validate() = %q, which is a raw validator error rather "+
					"than one of the errs sentinels", err)
			}
		})
	}
}

func TestValidModelsPassValidation(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T) any
	}{
		{name: "project", build: func(t *testing.T) any { return *newProject(t) }},
		{name: "task", build: func(t *testing.T) any { return *newTask(t) }},
		{name: "block", build: func(t *testing.T) any { return *newBlock(t) }},
		{
			name: "running block",
			build: func(t *testing.T) any {
				b := newBlock(t)
				b.TimestampEnd = time.Time{}
				return *b
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := val.Validate(test.build(t)); err != nil {
				t.Errorf("Validate() = %s, want nil", err)
			}
		})
	}
}
