package activeblock

import (
	"log/slog"
	"testing"

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

func TestNewIsEmpty(t *testing.T) {
	ab, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	if ab.GetKey() != KEY {
		t.Errorf("GetKey() = %q, want %q", ab.GetKey(), KEY)
	}
	if ab.HasActiveBlockKey() == true {
		t.Errorf("HasActiveBlockKey() = true on a fresh ActiveBlock")
	}
	if ab.HasPreviousBlockKey() == true {
		t.Errorf("HasPreviousBlockKey() = true on a fresh ActiveBlock")
	}
}

func TestSetActiveBlockKeyMovesTheOldKeyToPrevious(t *testing.T) {
	ab, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	ab.SetActiveBlockKey("block:1")

	if ab.GetActiveBlockKey() != "block:1" {
		t.Errorf("GetActiveBlockKey() = %q, want block:1", ab.GetActiveBlockKey())
	}
	if ab.HasPreviousBlockKey() == true {
		t.Errorf("the first active key was recorded as a previous key")
	}

	ab.SetActiveBlockKey("block:2")

	if ab.GetActiveBlockKey() != "block:2" {
		t.Errorf("GetActiveBlockKey() = %q, want block:2", ab.GetActiveBlockKey())
	}
	if ab.GetPreviousBlockKey() != "block:1" {
		t.Errorf("GetPreviousBlockKey() = %q, want block:1", ab.GetPreviousBlockKey())
	}
}

func TestClearActiveBlockKeyKeepsThePreviousKey(t *testing.T) {
	ab, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	ab.SetActiveBlockKey("block:1")
	ab.ClearActiveBlockKey()

	if ab.HasActiveBlockKey() == true {
		t.Errorf("HasActiveBlockKey() = true after clearing")
	}
	if ab.GetPreviousBlockKey() != "block:1" {
		t.Errorf("GetPreviousBlockKey() = %q, want block:1", ab.GetPreviousBlockKey())
	}
}

func TestClearActiveBlockKeyTwiceKeepsTheFirstPreviousKey(t *testing.T) {
	ab, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}

	ab.SetActiveBlockKey("block:1")
	ab.ClearActiveBlockKey()
	ab.ClearActiveBlockKey()

	if ab.GetPreviousBlockKey() != "block:1" {
		t.Errorf("GetPreviousBlockKey() = %q, want block:1", ab.GetPreviousBlockKey())
	}
}

func TestGetReturnsAnEmptyActiveBlockWhenNoneIsStored(t *testing.T) {
	db := newTestDatabase(t)

	ab, err := Get(db)
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if ab == nil {
		t.Fatalf("Get() = nil")
	}
	if ab.HasActiveBlockKey() == true {
		t.Errorf("Get() returned an ActiveBlock with a key on an empty database")
	}
}

func TestSetAndGetRoundTrip(t *testing.T) {
	db := newTestDatabase(t)

	stored, err := New()
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	stored.SetActiveBlockKey("block:1")
	stored.SetActiveBlockKey("block:2")

	if err = Set(db, stored); err != nil {
		t.Fatalf("Set() = %s", err)
	}

	loaded, err := Get(db)
	if err != nil {
		t.Fatalf("Get() = %s", err)
	}

	if loaded.GetActiveBlockKey() != "block:2" {
		t.Errorf("GetActiveBlockKey() = %q, want block:2", loaded.GetActiveBlockKey())
	}
	if loaded.GetPreviousBlockKey() != "block:1" {
		t.Errorf("GetPreviousBlockKey() = %q, want block:1", loaded.GetPreviousBlockKey())
	}
	if loaded.GetKey() != KEY {
		t.Errorf("GetKey() = %q, want %q", loaded.GetKey(), KEY)
	}
}
