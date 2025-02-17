package z

import (
  "os"
  "fmt"
  "strings"
  "encoding/json"
  "github.com/spf13/cobra"
)

func exportZeitJson(user string, entries []Entry) (string, error) {
  stringified, err := json.Marshal(entries)
  if err != nil {
    return "", err
  }

  return string(stringified), nil
}

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
    sinceTime, untilTime := ParseSinceUntil(since, until, listRange)

    var filteredEntries []Entry
    filteredEntries, err = GetFilteredEntries(entries, project, task, sinceTime, untilTime)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    var output string = ""
    switch(format) {
    case "zeit":
      output, err = exportZeitJson(user, filteredEntries)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    case "tyme":
      output, err = exportTymeJson(user, filteredEntries)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    default:
      fmt.Printf("%s specify an export format; see `zeit export --help` for more info\n", CharError)
      os.Exit(1)
    }

    fmt.Printf("%s\n", output)
    return
  },
}

func init() {
  rootCmd.AddCommand(exportCmd)
  exportCmd.Flags().StringVar(&format, "format", "zeit", "Format to export, possible values: zeit, tyme")
  exportCmd.Flags().StringVar(&since, "since", "", "Date/time to start the export from")
  exportCmd.Flags().StringVar(&until, "until", "", "Date/time to export until")
  exportCmd.Flags().StringVar(&listRange, "range", "", "Shortcut for --since and --until that accepts: " + strings.Join(Ranges(), ", "))
  exportCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be exported")
  exportCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be exported")

  flagName := "task"
  exportCmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    user := GetCurrentUser()
    entries, _ := database.ListEntries(user)
    _, tasks := listProjectsAndTasks(entries)
    return tasks, cobra.ShellCompDirectiveDefault
  })
}
