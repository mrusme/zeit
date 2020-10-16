package z

import (
  "os"
  "fmt"
  "github.com/spf13/cobra"
  "github.com/shopspring/decimal"
)

var statsCmd = &cobra.Command{
  Use:   "stats",
  Short: "Display activity statistics",
  Long: "Display statistics on all tracked activities.",
  Run: func(cmd *cobra.Command, args []string) {
    // user := GetCurrentUser()

    // entries, err := database.ListEntries(user)
    // if err != nil {
    //   fmt.Printf("%s %+v\n", CharError, err)
    //   os.Exit(1)
    // }

    // for _, entry := range entries {
    //   fmt.Printf("%s\n", entry.GetOutput())
    // }

    var cal Calendar

    var data = make(map[string]decimal.Decimal)

    data["Mo"], _ = decimal.NewFromString("15.00")
    data["Tu"], _ = decimal.NewFromString("4.0")
    data["We"], _ = decimal.NewFromString("10.0")
    data["Th"], _ = decimal.NewFromString("1.0")
    data["Fr"], _ = decimal.NewFromString("0.0")
    data["Sa"], _ = decimal.NewFromString("18.2")
    data["Su"], _ = decimal.NewFromString("1.0")

    out := cal.GetOutputForWeekCalendar(1, data)
    out2 := cal.GetOutputForWeekCalendar(2, data)

    fmt.Printf("%s\n", OutputAppendRight(out, out2, 10))

    return
  },
}

func init() {
  rootCmd.AddCommand(statsCmd)

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
