package z

import (
  "os"
  "fmt"
  "github.com/spf13/cobra"
)

var trackingCmd = &cobra.Command{
  Use:   "tracking",
  Short: "Currently tracking activity",
  Long: "Show currently tracking activity.",
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    runningEntryId, err := database.GetRunningEntryId(user)
    if err != nil {
      fmt.Printf("△ %+v\n", err)
      os.Exit(1)
    }

    if runningEntryId == "" {
      fmt.Printf("□ not running\n")
      os.Exit(1)
    }

    runningEntry, err := database.GetEntry(user, runningEntryId)
    if err != nil {
      fmt.Printf("△ %+v\n", err)
      os.Exit(1)
    }

    fmt.Printf(runningEntry.GetOutputForTrack(true, true))
    return
  },
}

func init() {
  rootCmd.AddCommand(trackingCmd)

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("△ %+v\n", err)
    os.Exit(1)
  }
}
