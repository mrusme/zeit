package z

import (
  "os"
  "fmt"
  "time"
  "github.com/spf13/cobra"
  "github.com/shopspring/decimal"
)

var listTotalTime bool

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

    totalHours := decimal.NewFromInt(0);
    for _, entry := range filteredEntries {
      duration := entry.Finish.Sub(entry.Begin)
      durationDec := decimal.NewFromFloat(duration.Hours())
      totalHours = totalHours.Add(durationDec)
      fmt.Printf("%s\n", entry.GetOutput())
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

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
