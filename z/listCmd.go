package z

import (
  "os"
  "fmt"
  "time"
  "github.com/spf13/cobra"
  "github.com/shopspring/decimal"
)

var listTotalTime bool
var listOnlyProjectsAndTasks bool

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

    if listOnlyProjectsAndTasks == true {
      var projectsAndTasks = make(map[string]map[string]bool)

      for _, filteredEntry := range filteredEntries {
        taskMap, ok := projectsAndTasks[filteredEntry.Project]

        if !ok {
          taskMap = make(map[string]bool)
          projectsAndTasks[filteredEntry.Project] = taskMap
        }

        taskMap[filteredEntry.Task] = true
        projectsAndTasks[filteredEntry.Project] = taskMap
      }

      for project, _ := range projectsAndTasks {
        fmt.Printf("%s %s\n", CharMore, project)

        for task, _ := range projectsAndTasks[project] {
          fmt.Printf("%*s└── %s\n", 1, " ", task)
        }
      }

      return
    }

    totalHours := decimal.NewFromInt(0);
    for _, entry := range filteredEntries {
      totalHours = totalHours.Add(entry.GetDuration())
      fmt.Printf("%s\n", entry.GetOutput(false))
    }

    if listTotalTime == true {
      fmt.Printf("\nTOTAL: %s H\n\n", totalHours.StringFixed(2))
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
  listCmd.Flags().BoolVar(&listTotalTime, "total", false, "Show total time of hours for listed activities")
  listCmd.Flags().BoolVar(&listOnlyProjectsAndTasks, "only-projects-and-tasks", false, "Only list projects and their tasks, no entries")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
