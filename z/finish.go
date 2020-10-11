package z

import (
  "log"
  "time"
  "github.com/spf13/cobra"
)

var finishCmd = &cobra.Command{
  Use:   "finish",
  Short: "Finish currently running tracker",
  Long: "Finishing a currently running tracker.",
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    runningEntryId, err := database.GetRunningEntryId(user)
    if err != nil {
      log.Fatal(err)
    }

    if runningEntryId == "" {
      log.Fatal("Tracker not running!")
    }

    runningEntry, err := database.GetEntry(user, runningEntryId)
    if err != nil {
      log.Fatal(err)
    }

    runningEntry.Finish = time.Now()

    entryId, err := database.FinishEntry(user, runningEntry)
    if err != nil {
      log.Fatal(err)
    }

    // entries, err := database.ListEntries()
    // if err != nil {
    //   log.Fatal(err)
    // }
    // fmt.Printf("%+v", entries)

    log.Printf("Finished entry with ID %s!\n", entryId)
    return
  },
}

func init() {
  rootCmd.AddCommand(finishCmd)
  finishCmd.Flags().StringVarP(&begin, "begin", "b", "", "Time the entry should begin at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
  finishCmd.Flags().StringVarP(&finish, "finish", "s", "", "Time the entry should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
  finishCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be assigned")
  finishCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be assigned")
  finishCmd.Flags().BoolVarP(&force, "force", "f", false, "Force begin finishing of a new task \neven though another one is still running \n(ONLY IF YOU KNOW WHAT YOU'RE DOING!)")

  var err error
  database, err = InitDatabase()
  if err != nil {
    log.Fatal(err)
  }
}
