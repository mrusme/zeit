package z

import (
  "os"
  "sort"
  "strings"
  "log"
  "errors"
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

      entry.SetIDFromDatabaseKey(key)

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

func (database *Database) EraseEntry(user string, id string) (error) {
  runningEntryId, err := database.GetRunningEntryId(user)
  if err != nil {
    return err
  }

  dberr := database.DB.Update(func(tx *buntdb.Tx) error {
    if runningEntryId == id {
      _, _, seterr := tx.Set(user + ":status:running", "", nil)
      if seterr != nil {
        return seterr
      }
    }

    _, delerr := tx.Delete(user + ":entry:" + id)
    if delerr != nil {
      return delerr
    }

    return nil
  })

  return dberr
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

func (database *Database) ListEntries(user string) ([]Entry, error) {
  var entries []Entry

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    tx.AscendKeys(user + ":entry:*", func(key, value string) bool {
      var entry Entry
      json.Unmarshal([]byte(value), &entry)

      entry.SetIDFromDatabaseKey(key)

      entries = append(entries, entry)
      return true
    })

    return nil
  })

  sort.Slice(entries, func(i, j int) bool { return entries[i].Begin.Before(entries[j].Begin) })
  return entries, dberr
}

func (database *Database) GetImportsSHA1List(user string) (map[string]string, error) {
  var sha1List = make(map[string]string)

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    value, err := tx.Get(user + ":imports:sha1", false)
    if err != nil {
      return nil
    }

    sha1Entries := strings.Split(value, ",")

    for _, sha1Entry := range sha1Entries {
      sha1EntrySplit := strings.Split(sha1Entry, ":")
      sha1 := sha1EntrySplit[0]
      id := sha1EntrySplit[1]
      sha1List[sha1] = id
    }

    return nil
  })

  return sha1List, dberr
}

func (database *Database) UpdateImportsSHA1List(user string, sha1List map[string]string) (error) {
    var sha1Entries []string

    for sha1, id := range sha1List {
      sha1Entries = append(sha1Entries, sha1 + ":" + id)
    }

    value := strings.Join(sha1Entries, ",")

    dberr := database.DB.Update(func(tx *buntdb.Tx) error {
      _, _, seterr := tx.Set(user + ":imports:sha1", value, nil)
      if seterr != nil {
        return seterr
      }

      return nil
    })

    return dberr
}

func (database *Database) UpdateProject(user string, projectName string, project Project) (error) {
  projectJson, jsonerr := json.Marshal(project)
  if jsonerr != nil {
    return jsonerr
  }

  projectId := GetIdFromName(projectName)

  dberr := database.DB.Update(func(tx *buntdb.Tx) error {
    _, _, sperr := tx.Set(user + ":project:" + projectId, string(projectJson), nil)
    if sperr != nil {
      return sperr
    }

    return nil
  })

  return dberr
}

func (database *Database) GetProject(user string, projectName string) (Project, error) {
  var project Project
  projectId := GetIdFromName(projectName)

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    value, err := tx.Get(user + ":project:" + projectId, false)
    if err != nil {
      return nil
    }

    json.Unmarshal([]byte(value), &project)

    return nil
  })

  return project, dberr
}

func (database *Database) UpdateTask(user string, taskName string, task Task) (error) {
  taskJson, jsonerr := json.Marshal(task)
  if jsonerr != nil {
    return jsonerr
  }

  taskId := GetIdFromName(taskName)

  dberr := database.DB.Update(func(tx *buntdb.Tx) error {
    _, _, sperr := tx.Set(user + ":task:" + taskId, string(taskJson), nil)
    if sperr != nil {
      return sperr
    }

    return nil
  })

  return dberr
}

func (database *Database) GetTask(user string, taskName string) (Task, error) {
  var task Task
  taskId := GetIdFromName(taskName)

  dberr := database.DB.View(func(tx *buntdb.Tx) error {
    value, err := tx.Get(user + ":task:" + taskId, false)
    if err != nil {
      return nil
    }

    json.Unmarshal([]byte(value), &task)

    return nil
  })

  return task, dberr
}
