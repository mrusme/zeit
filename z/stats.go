package z

import (
  "os"
  "fmt"
  "time"
  "strings"
  "github.com/spf13/cobra"
  // "github.com/shopspring/decimal"
  // "github.com/gookit/color"
)

var statsCmd = &cobra.Command{
  Use:   "stats",
  Short: "Display activity statistics",
  Long: "Display statistics on all tracked activities.",
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()

    entries, err := database.ListEntries(user)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    cal, _ := NewCalendar(entries)

    today := time.Now()
    month, weeknumber := GetISOWeekInMonth(today)
    month0 := month - 1
    weeknumber0 := weeknumber - 1
    thisWeek := cal.GetOutputForWeekCalendar(today, month0, weeknumber0)

    oneWeekAgo := today.AddDate(0, 0, -7)
    month, weeknumber = GetISOWeekInMonth(oneWeekAgo)
    month0 = month - 1
    weeknumber0 = weeknumber - 1
    previousWeek := cal.GetOutputForWeekCalendar(oneWeekAgo, month0, weeknumber0)


    fmt.Printf("\n%s\n\n", strings.ToUpper(today.Month().String()))
    fmt.Printf("%s\n\n\n", OutputAppendRight(thisWeek, previousWeek, 16))
    fmt.Printf("%s\n", cal.GetOutputForDistribution())

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
