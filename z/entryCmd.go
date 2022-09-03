package z

import (
  "os"
  "fmt"
  "time"
  "strings"
  "github.com/spf13/cobra"
)

var entryCmd = &cobra.Command{
  Use:   "entry ([flags]) [id]",
  Short: "Display or update activity",
  Long: "Display or update tracked activity.",
  Args: cobra.MaximumNArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    var id string
    if len(args) == 0 {
      entryList, err := database.ListEntries(user)
      if err != nil {
        fmt.Printf("%s %+v\n", CharError, err)
        os.Exit(1)
      }
      id = entryList[len(entryList) - 1].ID
    } else {
      id = args[0]
    }


    entry, err := database.GetEntry(user, id)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    if begin != "" || finish != "" || project != "" || notes != "" || task != "" {
      if begin != "" {
        entry.Begin, err = time.Parse(time.RFC3339, begin)
        if err != nil {
          fmt.Printf("%s %+v\n", CharError, err)
          os.Exit(1)
        }
      }

      if finish != "" {
        entry.Finish, err = time.Parse(time.RFC3339, finish)
        if err != nil {
          fmt.Printf("%s %+v\n", CharError, err)
          os.Exit(1)
        }
      }

      if project != "" {
        entry.Project = project
      }

      if task != "" {
        entry.Task = task
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
  entryCmd.Flags().StringVarP(&begin, "begin", "b", "", "Update date/time the activity began at\n\nUse RFC3339 format.")
  entryCmd.Flags().StringVarP(&finish, "finish", "s", "", "Update date/time the activity finished at\n\nUse RFC3339 format.")
  entryCmd.Flags().StringVarP(&project, "project", "p", "", "Update activity project")
  entryCmd.Flags().StringVarP(&notes, "notes", "n", "", "Update activity notes")
  entryCmd.Flags().StringVarP(&task, "task", "t", "", "Update activity task")
  entryCmd.Flags().BoolVar(&fractional, "decimal", false, "Show fractional hours in decimal format instead of minutes")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
