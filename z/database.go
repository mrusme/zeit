package z

import (
  "log"
  "encoding/json"
  "github.com/tidwall/buntdb"
  "github.com/google/uuid"
)

type Database struct {
  DB *buntdb.DB
}

func InitDatabase() (*Database, error) {
  db, err := buntdb.Open(":memory:") // TODO: Replace with file
  if err != nil {
    return nil, err
  }

  db.CreateIndex("task", "*", buntdb.IndexJSON("task"))
  db.CreateIndex("project", "*", buntdb.IndexJSON("project"))

  database := Database{db}
  return &database, nil
}

func (database *Database) NewID() (string) {
  id, err := uuid.NewRandom()
  if err != nil {
    log.Fatalln("Could not generate UUID: %+v", err)
  }
  return id.String()
}

func (database *Database) AddEntry(entry Entry) (string, error) {
  id := database.NewID()

  entryJson, jsonerr := json.Marshal(entry)
  if jsonerr != nil {
    return id, jsonerr
  }

  dberr := database.DB.Update(func(tx *buntdb.Tx) error {
    _, _, seterr := tx.Set(id, string(entryJson), nil)
    return seterr
  })

  return id, dberr
}

func (database *Database) ListEntries() ([]Entry, error) {
  var entries []Entry

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    tx.Ascend("task", func(key, value string) bool {
      var entry Entry
      json.Unmarshal([]byte(value), &entry)

      entry.ID = key

      entries = append(entries, entry)
      return true
    })

    return nil
  })

  return entries, dberr
}
