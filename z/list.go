package z

import (
  "os"
  "fmt"
  "github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
  Use:   "list",
  Short: "List activity",
  Long: "List all tracked activity.",
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    entries, err := database.ListEntries(user)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    for _, entry := range entries {
      fmt.Printf("%s\n", entry.GetOutput())
    }

    return
  },
}

func init() {
  rootCmd.AddCommand(listCmd)

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
