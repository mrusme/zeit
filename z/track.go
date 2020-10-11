package z

import (
  "fmt"
  "log"
  "github.com/spf13/cobra"
)

var database *Database
var begin string
var finish string
var project string
var task string

var trackCmd = &cobra.Command{
  Use:   "track",
  Short: "Tracking time",
  Long: "Add a new tracking entry, which can either be kept running until 'finish' is being called or parameterized to be a finished entry.",
  Run: func(cmd *cobra.Command, args []string) {
    newEntry, err := NewEntry("", begin, finish, project, task, GetCurrentUser())
    database.AddEntry(newEntry)
    entries, err := database.ListEntries()
    if err != nil {
      log.Fatal(err)
    }

    fmt.Printf("%+v", entries)
  },
}

func init() {
  rootCmd.AddCommand(trackCmd)
  trackCmd.Flags().StringVar(&begin, "begin", "", "Time the entry should begin at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now - 15 minutes), +1:30 (now plus 1.5h).")
  trackCmd.Flags().StringVar(&finish, "finish", "", "Time the entry should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now - 15 minutes), +1:30 (now plus 1.5h).\nMust be after --start time.")
  trackCmd.Flags().StringVar(&project, "project", "", "Project to be assigned")
  trackCmd.Flags().StringVar(&task, "task", "", "Task to be assigned")

  var err error
  database, err = InitDatabase()
  if err != nil {
    log.Fatal(err)
  }
}
