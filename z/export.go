package z

import (
  "os"
  "fmt"
  "time"
  "github.com/spf13/cobra"
)

var exportSince string
var exportUntil string

func exportTymeJson(user string, entries []*Entry) (string, error) {

    tyme := Tyme{}
    err := tyme.FromEntries(entries)
    if err != nil {
      return "", err
    }

    return tyme.Stringify(), nil
}

var exportCmd = &cobra.Command{
  Use:   "export ([flags])",
  Short: "Export tracked activities",
  Long: "Export tracked activities to various formats.",
  // Args: cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    var entries []Entry
    var err error

    user := GetCurrentUser()

    entries, err = database.ListEntries(user)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    var since time.Time
    var until time.Time

    if exportSince != "" {
      since, err = time.Parse(time.RFC3339, exportSince)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    }

    if exportUntil != "" {
      until, err = time.Parse(time.RFC3339, exportUntil)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    }

    var filteredEntries []*Entry
    filteredEntries, err = GetFilteredEntries(&entries, project, task, since, until)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    var output string = ""
    if formatTymeJson == true {
      output, err = exportTymeJson(user, filteredEntries)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    } else {
      fmt.Printf("%s specify an export format; see `zeit export --help` for more info\n", CharError)
      os.Exit(1)
    }

    fmt.Printf("%s\n", output)

    return
  },
}

func init() {
  rootCmd.AddCommand(exportCmd)
  exportCmd.Flags().BoolVar(&formatTymeJson, "tyme", false, "Export to Tyme 3 JSON")
  exportCmd.Flags().StringVar(&exportSince, "since", "", "Date/time to start the export from")
  exportCmd.Flags().StringVar(&exportUntil, "until", "", "Date/time to export until")
  exportCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be exported")
  exportCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be exported")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
