package project

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

func mustInsert(t *testing.T, db *database.Database, sid string) *Project {
	t.Helper()

	pj, err := InsertIfNone(db, "owner", sid)
	if err != nil {
		t.Fatalf("InsertIfNone(%q) = %s", sid, err)
	}

	return pj
}

func TestNew(t *testing.T) {
	pj, err := New("owner", "myproject")
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if pj.SID != "myproject" {
		t.Errorf("SID = %q, want myproject", pj.SID)
	}
	if pj.OwnerKey != "owner" {
		t.Errorf("OwnerKey = %q, want owner", pj.OwnerKey)
	}
	if pj.DisplayName != "Myproject" {
		t.Errorf("DisplayName = %q, want Myproject", pj.DisplayName)
	}
	if regexp.MustCompile(`^#[0-9A-F]{6}$`).MatchString(pj.Color) == false {
		t.Errorf("Color = %q, want a hex color", pj.Color)
	}
	if strings.HasPrefix(pj.GetKey(), "project:") == false {
		t.Errorf("GetKey() = %q, want a project: prefix", pj.GetKey())
	}
}

func TestNewProducesAValidProject(t *testing.T) {
	db := newTestDatabase(t)

	pj, err := New("owner", "myproject")
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if err = Set(db, pj); err != nil {
		t.Errorf("Set() = %s, want a freshly built project to validate", err)
	}
}

func TestSetValidates(t *testing.T) {
	db := newTestDatabase(t)

	tests := []struct {
		name    string
		mutate  func(pj *Project)
		wantErr bool
	}{
		{name: "unchanged", mutate: func(pj *Project) {}},
		{
			name:    "empty sid",
			mutate:  func(pj *Project) { pj.SID = "" },
			wantErr: true,
		},
		{
			name:    "invalid sid",
			mutate:  func(pj *Project) { pj.SID = "my project" },
			wantErr: true,
		},
		{
			name:    "oversized sid",
			mutate:  func(pj *Project) { pj.SID = strings.Repeat("a", 33) },
			wantErr: true,
		},
		{
			name:    "oversized display name",
			mutate:  func(pj *Project) { pj.DisplayName = strings.Repeat("a", 33) },
			wantErr: true,
		},
		{
			name:    "invalid color",
			mutate:  func(pj *Project) { pj.Color = "red" },
			wantErr: true,
		},
		{
			name:    "empty color",
			mutate:  func(pj *Project) { pj.Color = "" },
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pj, err := New("owner", "myproject")
			if err != nil {
				t.Fatalf("New() = %s", err)
			}

			test.mutate(pj)

			err = Set(db, pj)
			if test.wantErr == true && err == nil {
				t.Errorf("Set() = nil, want an error")
			}
			if test.wantErr == false && err != nil {
				t.Errorf("Set() = %s, want nil", err)
			}
		})
	}
}

func TestGetBySID(t *testing.T) {
	db := newTestDatabase(t)

	stored := mustInsert(t, db, "myproject")
	mustInsert(t, db, "otherproject")

	loaded, err := GetBySID(db, "myproject")
	if err != nil {
		t.Fatalf("GetBySID() = %s", err)
	}

	if loaded.GetKey() != stored.GetKey() {
		t.Errorf("GetBySID() = %q, want %q", loaded.GetKey(), stored.GetKey())
	}
}

func TestGetBySIDMissing(t *testing.T) {
	db := newTestDatabase(t)

	mustInsert(t, db, "myproject")

	if _, err := GetBySID(db, "nothere"); errors.Is(err, errs.ErrSIDNotFound) == false {
		t.Errorf("GetBySID() = %v, want ErrSIDNotFound", err)
	}
}

func TestGetBySIDOnAnEmptyDatabase(t *testing.T) {
	db := newTestDatabase(t)

	if _, err := GetBySID(db, "myproject"); errors.Is(
		err, errs.ErrSIDNotFound,
	) == false {
		t.Errorf("GetBySID() = %v, want ErrSIDNotFound", err)
	}
}

func TestGetByKey(t *testing.T) {
	db := newTestDatabase(t)

	stored := mustInsert(t, db, "myproject")

	loaded, err := Get(db, stored.GetKey())
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if loaded.SID != "myproject" {
		t.Errorf("Get() = %q, want myproject", loaded.SID)
	}
}

func TestGetByKeyMissing(t *testing.T) {
	db := newTestDatabase(t)

	if _, err := Get(db, "project:missing"); errors.Is(
		err, errs.ErrKeyNotFound,
	) == false {
		t.Errorf("Get() = %v, want ErrKeyNotFound", err)
	}
}

func TestInsertIfNoneIsIdempotent(t *testing.T) {
	db := newTestDatabase(t)

	first := mustInsert(t, db, "myproject")
	second := mustInsert(t, db, "myproject")

	if first.GetKey() != second.GetKey() {
		t.Errorf("InsertIfNone() created a second project for the same SID")
	}

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	if len(rows) != 1 {
		t.Errorf("List() returned %d projects, want 1", len(rows))
	}
}

func TestInsertIfNoneRejectsInvalidSIDs(t *testing.T) {
	db := newTestDatabase(t)

	if _, err := InsertIfNone(db, "owner", "my project"); err == nil {
		t.Errorf("InsertIfNone() = nil, want an error for an invalid SID")
	}
}

func TestList(t *testing.T) {
	db := newTestDatabase(t)

	mustInsert(t, db, "alpha")
	mustInsert(t, db, "beta")

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	if len(rows) != 2 {
		t.Fatalf("List() returned %d projects, want 2", len(rows))
	}

	for key, pj := range rows {
		if pj.GetKey() != key {
			t.Errorf("project %q carries key %q", key, pj.GetKey())
		}
	}
}

func TestListOnAnEmptyDatabase(t *testing.T) {
	db := newTestDatabase(t)

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}

	if len(rows) != 0 {
		t.Errorf("List() returned %d projects, want none", len(rows))
	}
}

func TestSetUpdatesInPlace(t *testing.T) {
	db := newTestDatabase(t)

	stored := mustInsert(t, db, "myproject")
	stored.DisplayName = "Renamed"
	stored.Color = "#123456"

	if err := Set(db, stored); err != nil {
		t.Fatalf("Set() = %s", err)
	}

	loaded, err := GetBySID(db, "myproject")
	if err != nil {
		t.Fatalf("GetBySID() = %s", err)
	}

	if loaded.DisplayName != "Renamed" || loaded.Color != "#123456" {
		t.Errorf("Set() = %+v, want the updated values", loaded)
	}

	rows, err := List(db)
	if err != nil {
		t.Fatalf("List() = %s", err)
	}
	if len(rows) != 1 {
		t.Errorf("List() returned %d projects, want 1", len(rows))
	}
}
