package z

import (
  "os"
  "fmt"
  "strings"
  "github.com/spf13/cobra"
)

var entryCmd = &cobra.Command{
  Use:   "entry ([flags]) [id]",
  Short: "Display or update activity",
  Long: "Display or update tracked activity.",
  Args: cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()
    id := args[0]

    entry, err := database.GetEntry(user, id)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    if begin != "" || finish != "" || project != "" || notes != "" || task != "" {
      tmpEntry, err := NewEntry(entry.ID, begin, finish, project, task, user)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }

      if begin != "" {
        entry.Begin = tmpEntry.Begin
      }

      if finish != "" {
        entry.Finish = tmpEntry.Finish
      }

      if project != "" {
        entry.Project = tmpEntry.Project
      }

      if task != "" {
        entry.Task = tmpEntry.Task
      }

      if notes != "" {
        entry.Notes = strings.Replace(notes, "\\n", "\n", -1)
      }

      _, err = database.UpdateEntry(user, entry)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
    }

    fmt.Printf("%s %s\n", CharInfo, entry.GetOutput(true))
    return
  },
}

func init() {
  rootCmd.AddCommand(entryCmd)
  entryCmd.Flags().StringVarP(&begin, "begin", "b", "", "Update time the activity began at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
  entryCmd.Flags().StringVarP(&finish, "finish", "s", "", "Update time the activity finished at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
  entryCmd.Flags().StringVarP(&project, "project", "p", "", "Update activity project")
  entryCmd.Flags().StringVarP(&notes, "notes", "n", "", "Update activity notes")
  entryCmd.Flags().StringVarP(&task, "task", "t", "", "Update activity task")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
