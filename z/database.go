package z

import (
  "os"
  "log"
  "errors"
  "strings"
  "encoding/json"
  "github.com/tidwall/buntdb"
  "github.com/google/uuid"
)

type Database struct {
  DB *buntdb.DB
}

func InitDatabase() (*Database, error) {
  dbfile, ok := os.LookupEnv("ZEIT_DB")
  if ok == false || dbfile == "" {
    return nil, errors.New("please `export ZEIT_DB` to the location the zeit database should be stored at")
  }

  db, err := buntdb.Open(dbfile)
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
    log.Fatalln("could not generate UUID: %+v", err)
  }
  return id.String()
}

func (database *Database) AddEntry(user string, entry Entry, setRunning bool) (string, error) {
  id := database.NewID()

  entryJson, jsonerr := json.Marshal(entry)
  if jsonerr != nil {
    return id, jsonerr
  }

  dberr := database.DB.Update(func(tx *buntdb.Tx) error {
    if setRunning == true {
      _, _, seterr := tx.Set(user + ":status:running", id, nil)
      if seterr != nil {
        return seterr
      }
    }
    _, _, seterr := tx.Set(user + ":entry:" + id, string(entryJson), nil)
    if seterr != nil {
      return seterr
    }

    return nil
  })

  return id, dberr
}

func (database *Database) GetEntry(user string, entryId string) (Entry, error) {
  var entry Entry

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    tx.AscendKeys(user + ":entry:" + entryId, func(key, value string) bool {
      json.Unmarshal([]byte(value), &entry)
      entry.ID = (strings.Split(key, ":"))[2]
      return true
    })

    return nil
  })

  return entry, dberr
}

func (database *Database) FinishEntry(user string, entry Entry) (string, error) {
  entryJson, jsonerr := json.Marshal(entry)
  if jsonerr != nil {
    return entry.ID, jsonerr
  }

  dberr := database.DB.Update(func(tx *buntdb.Tx) error {
    runningEntryId, grerr := tx.Get(user + ":status:running")
    if grerr != nil {
      return errors.New("no currently running entry found!")
    }

    if runningEntryId != entry.ID {
      return errors.New("specified entry is not currently running!")
    }

    _, _, srerr := tx.Set(user + ":status:running", "", nil)
    if srerr != nil {
      return srerr
    }

    _, _, seerr := tx.Set(user + ":entry:" + entry.ID, string(entryJson), nil)
    if seerr != nil {
      return seerr
    }

    return nil
  })

  return entry.ID, dberr
}

func (database *Database) GetRunningEntryId(user string) (string, error) {
  var runningId string = ""

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    tx.AscendKeys(user + ":status:running", func(key, value string) bool {
      runningId = value
      return true
    })

    return nil
  })

  return runningId, dberr
}

func (database *Database) ListEntries() ([]Entry, error) {
  var entries []Entry

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    tx.AscendKeys("*:entry:*", func(key, value string) bool {
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
