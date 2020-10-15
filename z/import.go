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
      fmt.Printf("%+v", entries)
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
