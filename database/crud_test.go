package database

import (
	"log/slog"
	"testing"

	"xn--gckvb8fzb.com/zeit/helpers/log"
)

type row struct {
	key   string
	Label string `json:"label"`
	Count int    `json:"count"`
}

func (r *row) SetKey(k string) { r.key = k }
func (r *row) GetKey() string  { return r.key }

func newTestDatabase(t *testing.T) *Database {
	t.Helper()

	db, err := New(log.New(slog.LevelError), "", false)
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	t.Cleanup(db.Close)

	return db
}

func TestUpsertAndGetBytes(t *testing.T) {
	db := newTestDatabase(t)

	if err := db.UpsertRowAsBytes("row:1", []byte("payload")); err != nil {
		t.Fatalf("UpsertRowAsBytes() = %s", err)
	}

	got, err := db.GetRowAsBytes("row:1")
	if err != nil {
		t.Fatalf("GetRowAsBytes() = %s", err)
	}

	if string(got) != "payload" {
		t.Errorf("GetRowAsBytes() = %q, want %q", got, "payload")
	}
}

func TestUpsertAndGetString(t *testing.T) {
	db := newTestDatabase(t)

	if err := db.UpsertRowAsString("row:1", "payload"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}

	got, err := db.GetRowAsString("row:1")
	if err != nil {
		t.Fatalf("GetRowAsString() = %s", err)
	}

	if got != "payload" {
		t.Errorf("GetRowAsString() = %q, want %q", got, "payload")
	}
}

func TestUpsertOverwrites(t *testing.T) {
	db := newTestDatabase(t)

	if err := db.UpsertRowAsString("row:1", "first"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}
	if err := db.UpsertRowAsString("row:1", "second"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}

	got, err := db.GetRowAsString("row:1")
	if err != nil {
		t.Fatalf("GetRowAsString() = %s", err)
	}

	if got != "second" {
		t.Errorf("GetRowAsString() = %q, want the second value", got)
	}
}

func TestUpsertAndGetStruct(t *testing.T) {
	db := newTestDatabase(t)

	stored := &row{Label: "first", Count: 3}
	stored.SetKey("row:1")

	if err := db.UpsertRowAsStruct(stored); err != nil {
		t.Fatalf("UpsertRowAsStruct() = %s", err)
	}

	loaded := new(row)
	if err := db.GetRowAsStruct("row:1", loaded); err != nil {
		t.Fatalf("GetRowAsStruct() = %s", err)
	}

	if loaded.Label != "first" || loaded.Count != 3 {
		t.Errorf("GetRowAsStruct() = %+v, want the stored values", loaded)
	}
	if loaded.GetKey() != "row:1" {
		t.Errorf("GetRowAsStruct() left the key as %q, want %q", loaded.GetKey(), "row:1")
	}
}

func TestGetMissingRow(t *testing.T) {
	db := newTestDatabase(t)

	_, err := db.GetRowAsBytes("row:missing")
	if err == nil {
		t.Fatalf("GetRowAsBytes() = nil error for a missing key")
	}

	if db.IsErrKeyNotFound(err) == false {
		t.Errorf("IsErrKeyNotFound(%v) = false, want true", err)
	}
}

func TestIsErrKeyNotFoundRejectsOtherErrors(t *testing.T) {
	db := newTestDatabase(t)

	if db.IsErrKeyNotFound(nil) == true {
		t.Errorf("IsErrKeyNotFound(nil) = true, want false")
	}
}

func TestDestroyRow(t *testing.T) {
	db := newTestDatabase(t)

	if err := db.UpsertRowAsString("row:1", "payload"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}

	if err := db.DestroyRow("row:1"); err != nil {
		t.Fatalf("DestroyRow() = %s", err)
	}

	_, err := db.GetRowAsBytes("row:1")
	if db.IsErrKeyNotFound(err) == false {
		t.Errorf("after DestroyRow() the lookup returned %v, want a missing key", err)
	}
}

func TestDestroyMissingRow(t *testing.T) {
	db := newTestDatabase(t)

	if err := db.DestroyRow("row:missing"); err != nil {
		t.Errorf("DestroyRow() = %s, want nil for a missing key", err)
	}
}

func TestGetPrefixedRows(t *testing.T) {
	db := newTestDatabase(t)

	wanted := map[string]string{
		"alpha:1": "a1",
		"alpha:2": "a2",
		"alpha:3": "a3",
	}
	for key, value := range wanted {
		if err := db.UpsertRowAsString(key, value); err != nil {
			t.Fatalf("UpsertRowAsString() = %s", err)
		}
	}
	if err := db.UpsertRowAsString("beta:1", "b1"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}

	got, err := db.GetPrefixedRowsAsBytes("alpha:")
	if err != nil {
		t.Fatalf("GetPrefixedRowsAsBytes() = %s", err)
	}

	if len(got) != len(wanted) {
		t.Fatalf("GetPrefixedRowsAsBytes() returned %d rows, want %d", len(got), len(wanted))
	}
	for key, value := range wanted {
		if string(got[key]) != value {
			t.Errorf("row %q = %q, want %q", key, got[key], value)
		}
	}
}

func TestGetPrefixedRowsAsStructSetsKeys(t *testing.T) {
	db := newTestDatabase(t)

	for _, key := range []string{"alpha:1", "alpha:2"} {
		stored := &row{Label: key}
		stored.SetKey(key)

		if err := db.UpsertRowAsStruct(stored); err != nil {
			t.Fatalf("UpsertRowAsStruct() = %s", err)
		}
	}
	if err := db.UpsertRowAsString("beta:1", `{"label":"beta"}`); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}

	rows := make(map[string]*row)
	if err := GetPrefixedRowsAsStruct(db, "alpha:", rows); err != nil {
		t.Fatalf("GetPrefixedRowsAsStruct() = %s", err)
	}

	if len(rows) != 2 {
		t.Fatalf("GetPrefixedRowsAsStruct() returned %d rows, want 2", len(rows))
	}
	for key, value := range rows {
		if value.GetKey() != key {
			t.Errorf("row %q has key %q", key, value.GetKey())
		}
		if value.Label != key {
			t.Errorf("row %q has label %q", key, value.Label)
		}
	}
}

func TestGetPrefixedRowsAsStructOnAnEmptyPrefix(t *testing.T) {
	db := newTestDatabase(t)

	rows := make(map[string]*row)
	if err := GetPrefixedRowsAsStruct(db, "nothing:", rows); err != nil {
		t.Fatalf("GetPrefixedRowsAsStruct() = %s", err)
	}

	if len(rows) != 0 {
		t.Errorf("GetPrefixedRowsAsStruct() returned %d rows, want none", len(rows))
	}
}

func TestGetAllRowsAsStruct(t *testing.T) {
	db := newTestDatabase(t)

	for _, key := range []string{"alpha:1", "beta:1"} {
		stored := &row{Label: key}
		stored.SetKey(key)

		if err := db.UpsertRowAsStruct(stored); err != nil {
			t.Fatalf("UpsertRowAsStruct() = %s", err)
		}
	}

	rows := make(map[string]*row)
	if err := GetAllRowsAsStruct(db, rows); err != nil {
		t.Fatalf("GetAllRowsAsStruct() = %s", err)
	}

	if len(rows) != 2 {
		t.Errorf("GetAllRowsAsStruct() returned %d rows, want 2", len(rows))
	}
}

func TestGetRowAsStructRejectsMalformedPayloads(t *testing.T) {
	db := newTestDatabase(t)

	if err := db.UpsertRowAsString("row:1", "not json"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}

	if err := db.GetRowAsStruct("row:1", new(row)); err == nil {
		t.Errorf("GetRowAsStruct() = nil error for a malformed payload")
	}
}

func TestNewRejectsAMissingReadOnlyDatabase(t *testing.T) {
	if _, err := New(log.New(slog.LevelError), t.TempDir(), true); err == nil {
		t.Errorf("New() = nil error for an empty read-only directory")
	}
}

func TestNewOnDisk(t *testing.T) {
	dir := t.TempDir()

	db, err := New(log.New(slog.LevelError), dir, false)
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if err = db.UpsertRowAsString("row:1", "payload"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}
	db.Close()

	reopened, err := New(log.New(slog.LevelError), dir, true)
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer reopened.Close()

	got, err := reopened.GetRowAsString("row:1")
	if err != nil {
		t.Fatalf("GetRowAsString() = %s", err)
	}

	if got != "payload" {
		t.Errorf("GetRowAsString() = %q, want %q", got, "payload")
	}
}

func TestReadOnlyDatabaseRejectsWrites(t *testing.T) {
	dir := t.TempDir()

	db, err := New(log.New(slog.LevelError), dir, false)
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	if err = db.UpsertRowAsString("row:1", "payload"); err != nil {
		t.Fatalf("UpsertRowAsString() = %s", err)
	}
	db.Close()

	reopened, err := New(log.New(slog.LevelError), dir, true)
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer reopened.Close()

	if err = reopened.UpsertRowAsString("row:2", "payload"); err == nil {
		t.Errorf("UpsertRowAsString() = nil error on a read-only database")
	}
}
