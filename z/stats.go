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

    var data = make(map[string][]Statistic)

    data["Mo"] = []Statistic {
      Statistic{ Hours: decimal.NewFromFloat(12.0), Project: "zeit", Color: color.FgRed.Render },
      Statistic{ Hours: decimal.NewFromFloat(3.5), Project: "blog", Color: color.FgGreen.Render },
    }
    data["Tu"] = []Statistic {
      Statistic{ Hours: decimal.NewFromFloat(2.25), Project: "zeit", Color: color.FgRed.Render },
      Statistic{ Hours: decimal.NewFromFloat(4.0), Project: "blog", Color: color.FgGreen.Render },
    }
    data["We"] = []Statistic {
      Statistic{ Hours: decimal.NewFromFloat(10.0), Project: "zeit", Color: color.FgRed.Render },
      Statistic{ Hours: decimal.NewFromFloat(1.5), Project: "blog", Color: color.FgGreen.Render },
    }
    data["Th"] = []Statistic {
      Statistic{ Hours: decimal.NewFromFloat(4.0), Project: "zeit", Color: color.FgRed.Render },
      Statistic{ Hours: decimal.NewFromFloat(4.5), Project: "blog", Color: color.FgGreen.Render },
    }
    data["Fr"] = []Statistic {
      Statistic{ Hours: decimal.NewFromFloat(0.5), Project: "zeit", Color: color.FgRed.Render },
      Statistic{ Hours: decimal.NewFromFloat(3.5), Project: "blog", Color: color.FgGreen.Render },
    }
    data["Sa"] = []Statistic {
      Statistic{ Hours: decimal.NewFromFloat(1.0), Project: "zeit", Color: color.FgRed.Render },
      Statistic{ Hours: decimal.NewFromFloat(1.0), Project: "blog", Color: color.FgGreen.Render },
    }
    data["Su"] = []Statistic {
      Statistic{ Hours: decimal.NewFromFloat(10.0), Project: "zeit", Color: color.FgRed.Render },
      Statistic{ Hours: decimal.NewFromFloat(0.5), Project: "blog", Color: color.FgGreen.Render },
    }

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
