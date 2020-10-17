package z

import (
  "os"
  "fmt"
  "time"
  "github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
  Use:   "list",
  Short: "List activities",
  Long: "List all tracked activities.",
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    entries, err := database.ListEntries(user)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    var sinceTime time.Time
    var untilTime time.Time

    if since != "" {
      sinceTime, err = time.Parse(time.RFC3339, since)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    }

    if until != "" {
      untilTime, err = time.Parse(time.RFC3339, until)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    }

    var filteredEntries []Entry
    filteredEntries, err = GetFilteredEntries(entries, project, task, sinceTime, untilTime)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    for _, entry := range filteredEntries {
      fmt.Printf("%s\n", entry.GetOutput())
    }

    return
  },
}

func init() {
  rootCmd.AddCommand(listCmd)
  listCmd.Flags().StringVar(&since, "since", "", "Date/time to start the list from")
  listCmd.Flags().StringVar(&until, "until", "", "Date/time to list until")
  listCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be listed")
  listCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be listed")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
