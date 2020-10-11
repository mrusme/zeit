package z

import (
  "log"
  "github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
  Use:   "track",
  Short: "Tracking time",
  Long: "Add a new tracking entry, which can either be kept running until 'finish' is being called or parameterized to be a finished entry.",
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    runningEntryId, err := database.GetRunningEntryId(user)
    if err != nil {
      log.Fatal(err)
    }

    if runningEntryId != "" {
      log.Fatal("A task is already running. Please finish that before beginning to track a new task!")
    }

    newEntry, err := NewEntry("", begin, finish, project, task, user)
    if err != nil {
      log.Fatal(err)
    }

    entryId, err := database.AddEntry(user, newEntry, true)
    if err != nil {
      log.Fatal(err)
    }

    // entries, err := database.ListEntries()
    // if err != nil {
    //   log.Fatal(err)
    // }
    // fmt.Printf("%+v", entries)

    log.Printf("Added new entry with ID %s!\n", entryId)
    return
  },
}

func init() {
  rootCmd.AddCommand(trackCmd)
  trackCmd.Flags().StringVarP(&begin, "begin", "b", "", "Time the entry should begin at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
  trackCmd.Flags().StringVarP(&finish, "finish", "s", "", "Time the entry should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
  trackCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be assigned")
  trackCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be assigned")
  trackCmd.Flags().BoolVarP(&force, "force", "f", false, "Force begin tracking of a new task \neven though another one is still running \n(ONLY IF YOU KNOW WHAT YOU'RE DOING!)")

  var err error
  database, err = InitDatabase()
  if err != nil {
    log.Fatal(err)
  }
}
