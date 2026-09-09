package task

import (
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"xn--gckvb8fzb.com/zeit/database"
	"xn--gckvb8fzb.com/zeit/errs"
	"xn--gckvb8fzb.com/zeit/helpers/log"
)

func newTestDatabase(t *testing.T) *database.Database {
	t.Helper()

	db, err := database.New(log.New(slog.LevelError), "", false)
	if err != nil {
		t.Fatalf("database.New() = %s", err)
	}

	t.Cleanup(db.Close)

	return db
}

func mustInsert(t *testing.T, db *database.Database, projectSID, sid string) *Task {
	t.Helper()

	tk, err := InsertIfNone(db, "owner", projectSID, sid)
	if err != nil {
		t.Fatalf("InsertIfNone(%q, %q) = %s", projectSID, sid, err)
	}

	return tk
}

func TestNew(t *testing.T) {
	tk, err := New("owner", "myproject", "mytask")
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if tk.SID != "mytask" {
		t.Errorf("SID = %q, want mytask", tk.SID)
	}
	if tk.ProjectSID != "myproject" {
		t.Errorf("ProjectSID = %q, want myproject", tk.ProjectSID)
	}
	if tk.OwnerKey != "owner" {
		t.Errorf("OwnerKey = %q, want owner", tk.OwnerKey)
	}
	if tk.DisplayName != "Mytask" {
		t.Errorf("DisplayName = %q, want Mytask", tk.DisplayName)
	}
	if regexp.MustCompile(`^#[0-9A-F]{6}$`).MatchString(tk.Color) == false {
		t.Errorf("Color = %q, want a hex color", tk.Color)
	}
	if strings.HasPrefix(tk.GetKey(), "task:") == false {
		t.Errorf("GetKey() = %q, want a task: prefix", tk.GetKey())
	}
}

func TestSetValidates(t *testing.T) {
	db := newTestDatabase(t)

	tests := []struct {
		name    string
		mutate  func(tk *Task)
		wantErr bool
	}{
		{name: "unchanged", mutate: func(tk *Task) {}},
		{
			name:    "empty sid",
			mutate:  func(tk *Task) { tk.SID = "" },
			wantErr: true,
		},
		{
			name:    "empty project sid",
			mutate:  func(tk *Task) { tk.ProjectSID = "" },
			wantErr: true,
		},
		{
			name:    "invalid sid",
			mutate:  func(tk *Task) { tk.SID = "my task" },
			wantErr: true,
		},
		{
			name:    "oversized display name",
			mutate:  func(tk *Task) { tk.DisplayName = strings.Repeat("a", 33) },
			wantErr: true,
		},
		{
			name:    "invalid color",
			mutate:  func(tk *Task) { tk.Color = "red" },
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tk, err := New("owner", "myproject", "mytask")
			if err != nil {
				t.Fatalf("New() = %s", err)
			}

			test.mutate(tk)

			err = Set(db, tk)
			if test.wantErr == true && err == nil {
				t.Errorf("Set() = nil, want an error")
			}
			if test.wantErr == false && err != nil {
				t.Errorf("Set() = %s, want nil", err)
			}
		})
	}
}

func TestGetBySIDIsScopedToTheProject(t *testing.T) {
	db := newTestDatabase(t)

	alpha := mustInsert(t, db, "alpha", "shared")
	beta := mustInsert(t, db, "beta", "shared")

	loaded, err := GetBySID(db, "alpha", "shared")
	if err != nil {
		t.Fatalf("GetBySID() = %s", err)
	}

	if loaded.GetKey() != alpha.GetKey() {
		t.Errorf("GetBySID(alpha) returned the task of another project")
	}

	loaded, err = GetBySID(db, "beta", "shared")
	if err != nil {
		t.Fatalf("GetBySID() = %s", err)
	}

	if loaded.GetKey() != beta.GetKey() {
		t.Errorf("GetBySID(beta) returned the task of another project")
	}
}

func TestGetBySIDMissing(t *testing.T) {
	db := newTestDatabase(t)

	mustInsert(t, db, "alpha", "one")

	tests := []struct {
		projectSID string
		sid        string
	}{
		{projectSID: "alpha", sid: "two"},
		{projectSID: "beta", sid: "one"},
		{projectSID: "", sid: ""},
	}

	for _, test := range tests {
		if _, err := GetBySID(db, test.projectSID, test.sid); errors.Is(
			err, errs.ErrSIDNotFound,
		) == false {
			t.Errorf("GetBySID(%q, %q) = %v, want ErrSIDNotFound",
				test.projectSID, test.sid, err)
		}
	}
}

func TestGetByKey(t *testing.T) {
	db := newTestDatabase(t)

	stored := mustInsert(t, db, "alpha", "one")

	loaded, err := Get(db, stored.GetKey())
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if loaded.SID != "one" || loaded.ProjectSID != "alpha" {
		t.Errorf("Get() = %s/%s, want alpha/one", loaded.ProjectSID, loaded.SID)
	}
}

func TestGetByKeyMissing(t *testing.T) {
	db := newTestDatabase(t)

	if _, err := Get(db, "task:missing"); errors.Is(
		err, errs.ErrKeyNotFound,
	) == false {
		t.Errorf("Get() = %v, want ErrKeyNotFound", err)
	}
}

func TestInsertIfNoneIsScopedToTheProject(t *testing.T) {
	db := newTestDatabase(t)

	alpha := mustInsert(t, db, "alpha", "shared")
	beta := mustInsert(t, db, "beta", "shared")

	if alpha.GetKey() == beta.GetKey() {
		t.Errorf("InsertIfNone() reused a task across two projects")
	}

	again := mustInsert(t, db, "alpha", "shared")
	if again.GetKey() != alpha.GetKey() {
		t.Errorf("InsertIfNone() created a second task for the same project and SID")
	}

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}
	if len(rows) != 2 {
		t.Errorf("List() returned %d tasks, want 2", len(rows))
	}
}

func TestListForProjectSID(t *testing.T) {
	db := newTestDatabase(t)

	mustInsert(t, db, "alpha", "one")
	mustInsert(t, db, "alpha", "two")
	mustInsert(t, db, "beta", "one")

	rows, err := ListForProjectSID(db, "alpha")
	if err != nil {
		t.Fatalf("ListForProjectSID() = %s", err)
	}

	if len(rows) != 2 {
		t.Fatalf("ListForProjectSID() returned %d tasks, want 2", len(rows))
	}

	for _, tk := range rows {
		if tk.ProjectSID != "alpha" {
			t.Errorf("ListForProjectSID(alpha) returned a task of %q", tk.ProjectSID)
		}
	}
}

func TestListForProjectSIDWithoutMatches(t *testing.T) {
	db := newTestDatabase(t)

	mustInsert(t, db, "alpha", "one")

	rows, err := ListForProjectSID(db, "nothere")
	if err != nil {
		t.Fatalf("ListForProjectSID() = %s", err)
	}

	if len(rows) != 0 {
		t.Errorf("ListForProjectSID() returned %d tasks, want none", len(rows))
	}
}

func TestGroupByProjectSID(t *testing.T) {
	db := newTestDatabase(t)

	mustInsert(t, db, "alpha", "one")
	mustInsert(t, db, "alpha", "two")
	mustInsert(t, db, "beta", "one")

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	grouped := GroupByProjectSID(rows)

	if len(grouped) != 2 {
		t.Fatalf("GroupByProjectSID() returned %d projects, want 2", len(grouped))
	}
	if len(grouped["alpha"]) != 2 {
		t.Errorf("alpha holds %d tasks, want 2", len(grouped["alpha"]))
	}
	if len(grouped["beta"]) != 1 {
		t.Errorf("beta holds %d tasks, want 1", len(grouped["beta"]))
	}
	if len(grouped["nothere"]) != 0 {
		t.Errorf("an absent project returned %d tasks, want none", len(grouped["nothere"]))
	}

	for projectSID, tasks := range grouped {
		for key, tk := range tasks {
			if tk.ProjectSID != projectSID {
				t.Errorf("task %q is grouped under %q but belongs to %q",
					key, projectSID, tk.ProjectSID)
			}
			if tk.GetKey() != key {
				t.Errorf("task %q carries key %q", key, tk.GetKey())
			}
		}
	}
}

func TestGroupByProjectSIDMatchesListForProjectSID(t *testing.T) {
	db := newTestDatabase(t)

	mustInsert(t, db, "alpha", "one")
	mustInsert(t, db, "alpha", "two")
	mustInsert(t, db, "beta", "one")

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	grouped := GroupByProjectSID(rows)

	for _, projectSID := range []string{"alpha", "beta", "nothere"} {
		listed, err := ListForProjectSID(db, projectSID)
		if err != nil {
			t.Fatalf("ListForProjectSID() = %s", err)
		}

		if len(listed) != len(grouped[projectSID]) {
			t.Errorf("%q: ListForProjectSID() has %d tasks, grouping has %d",
				projectSID, len(listed), len(grouped[projectSID]))
		}
		for key := range listed {
			if grouped[projectSID][key] == nil {
				t.Errorf("%q: task %q is missing from the grouping", projectSID, key)
			}
		}
	}
}

func TestGroupByProjectSIDOnAnEmptyMap(t *testing.T) {
	grouped := GroupByProjectSID(map[string]*Task{})

	if len(grouped) != 0 {
		t.Errorf("GroupByProjectSID() = %v, want an empty grouping", grouped)
	}
}
