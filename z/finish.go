package z

import (
  "os"
  "log"
  "fmt"
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
      fmt.Printf("□ not running\n")
      os.Exit(-1)
    }

    runningEntry, err := database.GetEntry(user, runningEntryId)
    if err != nil {
      log.Fatal(err)
    }

    tmpEntry, err := NewEntry(runningEntry.ID, begin, finish, project, task, user)
    if err != nil {
      log.Fatal(err)
    }

    if begin != "" {
      runningEntry.Begin = tmpEntry.Begin
    }

    if finish != "" {
      runningEntry.Finish = tmpEntry.Finish
    } else {
      runningEntry.Finish = time.Now()
    }

    if project != "" {
      runningEntry.Project = tmpEntry.Project
    }

    if task != "" {
      runningEntry.Task = tmpEntry.Task
    }

    _, err = database.FinishEntry(user, runningEntry)
    if err != nil {
      log.Fatal(err)
    }

    fmt.Printf(runningEntry.GetOutputForFinish())
    return
  },
}

func init() {
  rootCmd.AddCommand(finishCmd)
  finishCmd.Flags().StringVarP(&begin, "begin", "b", "", "Time the entry should begin at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
  finishCmd.Flags().StringVarP(&finish, "finish", "s", "", "Time the entry should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
  finishCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be assigned")
  finishCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be assigned")

  var err error
  database, err = InitDatabase()
  if err != nil {
    log.Fatal(err)
  }
}
