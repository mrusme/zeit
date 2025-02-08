package z

import (
  "os"
  "fmt"
  "time"
  "github.com/spf13/cobra"
)

var finishCmd = &cobra.Command{
  Use:   "finish",
  Short: "Finish currently running activity",
  Long: "Finishing tracking of currently running activity.",
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    runningEntryId, err := database.GetRunningEntryId(user)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    if runningEntryId == "" {
      fmt.Printf("%s not running\n", CharFinish)
      os.Exit(1)
    }

    runningEntry, err := database.GetEntry(user, runningEntryId)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    tmpEntry, err := NewEntry(runningEntry.ID, begin, finish, project, task, user)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
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

    if notes != "" {
      runningEntry.Notes = fmt.Sprintf("%s\n%s", runningEntry.Notes, notes)
    }

    if runningEntry.Task != "" {
      task, err := database.GetTask(user, runningEntry.Task)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }

      if task.GitRepository != "" && task.GitRepository != "-" {
        stdout, stderr, err := GetGitLog(task.GitRepository, runningEntry.Begin, runningEntry.Finish)
        if err != nil {
          fmt.Printf("%s %+v\n", CharError, err)
          os.Exit(1)
        }

        if stderr == "" {
          runningEntry.Notes = fmt.Sprintf("%s\n%s", runningEntry.Notes, stdout)
        } else {
          fmt.Printf("%s notes were not imported: %+v\n", CharError, stderr)
        }
      }
    }

    if runningEntry.IsFinishedAfterBegan() == false {
      fmt.Printf("%s %+v\n", CharError, "beginning time of tracking cannot be after finish time")
      os.Exit(1)
    }

    _, err = database.FinishEntry(user, runningEntry)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    fmt.Printf(runningEntry.GetOutputForFinish())
    return
  },
}

func init() {
  rootCmd.AddCommand(finishCmd)
  finishCmd.Flags().StringVarP(&begin, "begin", "b", "", "Time the activity should begin at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
  finishCmd.Flags().StringVarP(&finish, "finish", "s", "", "Time the activity should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
  finishCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be assigned")
  finishCmd.Flags().StringVarP(&notes, "notes", "n", "", "Activity notes")
  finishCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be assigned")
}
