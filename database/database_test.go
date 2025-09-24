package database

import (
	"log/slog"
	"strconv"
	"testing"

	"github.com/mrusme/zeit/helpers/log"
)

type TestData struct {
	key string
	ID  string
}

func (td *TestData) SetKey(k string) {
	td.key = k
}

func (td *TestData) GetKey() string {
	return td.key
}

func TestGetPrefixedRowsAsStruct(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	db, err := New(logger, "")
	if err != nil {
		t.Errorf("New database failed: %s\n", err)
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		is := strconv.Itoa(i)
		testdata := new(TestData)
		testdata.SetKey("testkey")
		testdata.ID = is
		err = db.UpsertRowAsStruct(testdata)
		if err != nil {
			t.Errorf("Upsert failed: %s\n", err)
		}
	}

	var rows map[string]*TestData = make(map[string]*TestData)
	err = GetPrefixedRowsAsStruct(db, "testkey", rows)
	if err != nil {
		t.Errorf("GetPrefixedRowsAsStruct failed: %s\n", err)
	}

	if len(rows) < 1 {
		t.Errorf("No results retrieved\n")
	}

	for key, value := range rows {
		t.Logf("Retrieved '%s': %v\n", key, value)
	}
}

