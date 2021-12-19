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

    weekMinus0 := time.Now()
    monthMinus0, weeknumberMinus0 := GetISOWeekInMonth(weekMinus0)
    monthMinus00 := monthMinus0 - 1
    weeknumberMinus00 := weeknumberMinus0 - 1
    thisWeek := cal.GetOutputForWeekCalendar(weekMinus0, monthMinus00, weeknumberMinus00)

    weekMinus1 := weekMinus0.AddDate(0, 0, -7)
    monthMinus1, weeknumberMinus1 := GetISOWeekInMonth(weekMinus1)
    monthMinus10 := monthMinus1 - 1
    weeknumberMinus10 := weeknumberMinus1 - 1
    previousWeek := cal.GetOutputForWeekCalendar(weekMinus1, monthMinus10, weeknumberMinus10)

    if monthMinus00 == monthMinus10 {
      fmt.Printf("\n%s\n\n", strings.ToUpper(weekMinus0.Month().String()))
    } else {
      fmt.Printf("\n%s / %s\n\n", strings.ToUpper(weekMinus0.Month().String()), strings.ToUpper(weekMinus1.Month().String()))
    }
    fmt.Printf("%s\n\n\n", OutputAppendRight(thisWeek, previousWeek, 16))
    fmt.Printf("%s\n", cal.GetOutputForDistribution())

    return
  },
}

func init() {
  rootCmd.AddCommand(statsCmd)
  statsCmd.Flags().BoolVar(&fractional, "decimal", false, "Show fractional hours in decimal format instead of minutes")
  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
