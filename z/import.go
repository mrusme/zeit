package z

import (
  "os"
  "fmt"
  "time"
  "github.com/spf13/cobra"
  "github.com/cnf/structhash"
)

var formatTymeJson bool

func importTymeJson(user string, file string) ([]Entry, error) {
    var entries []Entry

    tyme := Tyme{}
    tyme.Load(file)

    for _, tymeEntry := range tyme.Data {
      tymeEntrySHA1 := structhash.Sha1(tymeEntry, 1)
      fmt.Printf("%x\n", tymeEntrySHA1)

      tymeStart, err := time.Parse("2006-01-02T15:04:05-07:00", tymeEntry.Start)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        continue
      }

      tymeEnd, err := time.Parse("2006-01-02T15:04:05-07:00", tymeEntry.End)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        continue
      }

      entry, err := NewEntry("", "", "", tymeEntry.Project, tymeEntry.Task, user)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        continue
      }

      entry.Begin = tymeStart
      entry.Finish = tymeEnd

      entry.SHA1 = fmt.Sprintf("%x", tymeEntrySHA1)

      entries = append(entries, entry)
    }

    return entries, nil
}

var importCmd = &cobra.Command{
  Use:   "import ([flags]) [file]",
  Short: "Import tracked activities",
  Long: "Import tracked activities from various sources.",
  Args: cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    var entries []Entry
    var err error

    user := GetCurrentUser()

    if formatTymeJson == true {
      entries, err = importTymeJson(user, args[0])
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    } else {
      fmt.Printf("%s specify an import format; see `zeit import --help` for more info\n", CharError)
      os.Exit(1)
    }

    sha1List, sha1Err := database.GetImportsSHA1List(user)
    if sha1Err != nil {
        fmt.Printf("%s %+v\n", CharError, sha1Err)
        os.Exit(1)
    }

    for _, entry := range entries {
      if id, ok := sha1List[entry.SHA1]; ok {
        fmt.Printf("%s %s was previously imported as %s; not importing again\n", CharInfo, entry.SHA1, id)
        continue
      }

      importedId, err := database.AddEntry(user, entry, false)
      if err != nil {
        fmt.Printf("%s %s could not be imported: %+v\n", CharError, entry.SHA1, err)
        continue
      }

      fmt.Printf("%s %s was imported as %s\n", CharInfo, entry.SHA1, importedId)
      sha1List[entry.SHA1] = importedId
    }

    err = database.UpdateImportsSHA1List(user, sha1List)
    if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
    }

    return
  },
}

func init() {
  rootCmd.AddCommand(importCmd)
  importCmd.Flags().BoolVar(&formatTymeJson, "tyme", false, "Import from Tyme 3 JSON export")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
