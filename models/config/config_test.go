package config

import (
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"xn--gckvb8fzb.com/zeit/database"
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

func TestNewGeneratesAUserKey(t *testing.T) {
	cfg, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if cfg.GetKey() != KEY {
		t.Errorf("GetKey() = %q, want %q", cfg.GetKey(), KEY)
	}
	if _, err = uuid.Parse(cfg.UserKey); err != nil {
		t.Errorf("UserKey = %q, which is not a UUID: %s", cfg.UserKey, err)
	}
}

func TestNewGeneratesDistinctUserKeys(t *testing.T) {
	first, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	second, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if first.UserKey == second.UserKey {
		t.Errorf("New() returned the same user key twice")
	}
}

func TestGetDoesNotPersist(t *testing.T) {
	db := newTestDatabase(t)

	first, err := Get(db)
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	second, err := Get(db)
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if first.UserKey == second.UserKey {
		t.Errorf("Get() returned a stable user key without anything being stored")
	}
}

func TestInsertIfNonePersistsTheFirstUserKey(t *testing.T) {
	db := newTestDatabase(t)

	first, err := InsertIfNone(db)
	if err != nil {
		t.Fatalf("InsertIfNone() = %s", err)
	}

	second, err := InsertIfNone(db)
	if err != nil {
		t.Fatalf("InsertIfNone() = %s", err)
	}

	if first.UserKey != second.UserKey {
		t.Errorf("InsertIfNone() = %q then %q, want a stable user key",
			first.UserKey, second.UserKey)
	}

	loaded, err := Get(db)
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if loaded.UserKey != first.UserKey {
		t.Errorf("Get() = %q, want the persisted %q", loaded.UserKey, first.UserKey)
	}
}

func TestSetOverwritesTheStoredConfig(t *testing.T) {
	db := newTestDatabase(t)

	stored, err := InsertIfNone(db)
	if err != nil {
		t.Fatalf("InsertIfNone() = %s", err)
	}

	replacement, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	if err = Set(db, replacement); err != nil {
		t.Fatalf("Set() = %s", err)
	}

	loaded, err := Get(db)
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if loaded.UserKey == stored.UserKey {
		t.Errorf("Set() did not replace the stored config")
	}
	if loaded.UserKey != replacement.UserKey {
		t.Errorf("Get() = %q, want %q", loaded.UserKey, replacement.UserKey)
	}
}
