package z

import (
  "os"
  "fmt"
  // "time"
  "github.com/spf13/cobra"
)

func exportTymeJson(user string, entries []Entry) (string, error) {

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

    var output string = ""
    if formatTymeJson == true {
      output, err = exportTymeJson(user, entries)
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

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
