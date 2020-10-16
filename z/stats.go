package z

import (
  "os"
  "fmt"
  "github.com/spf13/cobra"
  "github.com/shopspring/decimal"
  "github.com/gookit/color"
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

    buff := cal.GetTuiBufferForWeekCalendar(1, data)
    buff2 := cal.GetTuiBufferForWeekCalendar(2, data)

    r := []rune(color.FgLightWhite.Render("A"))
    for _, bla := range r {
      fmt.Printf("Char: %c", bla)
    }

    tui := Tui{}
    // tui.Init()
    fmt.Printf("%s\n", tui.Render(100, 10))
    tui.Merge(buff, 0, 0)
    tui.Merge(buff2, 0, 48)
    fmt.Printf("---\n")
    fmt.Printf("%s\n", tui.Render(100, 10))

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
