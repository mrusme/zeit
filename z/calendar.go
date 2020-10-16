package z

import (
  "fmt"
  "time"
  "github.com/gookit/color"
  "github.com/shopspring/decimal"
)

type Statistic struct {
  Hours decimal.Decimal
  Project string
  Color (func(...interface {}) string)
}

type Calendar struct {

}

func GetOutputBoxForNumber(number int, clr (func(...interface {}) string) ) (string) {
  switch(number) {
  case 0: return clr("  ")
  case 1: return clr(" ▄")
  case 2: return clr("▄▄")
  case 3: return clr("▄█")
  case 4: return clr("██")
  }

  return clr("  ")
}

func GetOutputBarForHours(hours decimal.Decimal, stats []Statistic) ([]string) {
  var bar = []string{
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
    color.FgGray.Render("····"),
  }

  hoursInt := int((hours.Round(0)).IntPart())
  rest := ((hours.Round(0)).Mod(decimal.NewFromInt(4))).Round(0)
  restInt := int(rest.IntPart())

  divisible := hoursInt - restInt
  fullparts := divisible / 4

  colorsFull := make(map[int](func(...interface {}) string))
  colorsFullIdx := 0

  colorFraction := color.FgWhite.Render
  colorFractionPrevAmount := 0.0

  for _, stat := range stats {
    statHoursInt := int((stat.Hours.Round(0)).IntPart())
    statRest := (stat.Hours.Round(0)).Mod(decimal.NewFromInt(4))
    statRestFloat, _ := statRest.Float64()

    if statRestFloat > colorFractionPrevAmount {
      colorFractionPrevAmount = statRestFloat
      colorFraction = stat.Color
    }

    fullColoredParts := int(statHoursInt / 4)
    for i := 0; i < fullColoredParts; i++ {
      colorsFull[colorsFullIdx] = stat.Color
      colorsFullIdx++
    }
  }

  iColor := 0
  for i := (len(bar) - 1); i > (len(bar) - 1 - fullparts); i-- {
    if iColor < colorsFullIdx {
      bar[i] = " " + GetOutputBoxForNumber(4, colorsFull[iColor]) + " "
      iColor++
    } else {
      bar[i] = " " + GetOutputBoxForNumber(4, color.FgWhite.Render) + " "
    }
  }

  if(restInt > 0) {
    bar[(len(bar) - 1 - fullparts)] = " " + GetOutputBoxForNumber(restInt, colorFraction) + " "
  }

  return bar
}

func (calendar *Calendar) GetCalendarWeek(timestamp time.Time) (int) {
  var _, cw = timestamp.ISOWeek()
  return cw
}

func (calendar *Calendar) GetOutputForWeekCalendar(cw int, data map[string][]Statistic) (string) {
  var output string = ""
  var bars [][]string
  var totalHours = decimal.NewFromInt(0)

  var days = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
  for _, day := range days {
    var dayHours = decimal.NewFromInt(0)

    for _, stat := range data[day] {
      dayHours = dayHours.Add(stat.Hours)
      totalHours = totalHours.Add(stat.Hours)
    }

    bar := GetOutputBarForHours(dayHours, data[day])
    bars = append(bars, bar)
  }

  output = fmt.Sprintf("CW %02d                    %s H\n", cw, totalHours.StringFixed(2))
  for row := 0; row < len(bars[0]); row++ {
    output = fmt.Sprintf("%s%2d │", output, ((6 - row) * 4))
    for col := 0; col < len(bars); col++ {
      output = fmt.Sprintf("%s%s", output, bars[col][row])
    }
    output = fmt.Sprintf("%s\n", output)
  }
  output = fmt.Sprintf("%s   └────────────────────────────\n     %s  %s  %s  %s  %s  %s  %s\n",
    output, days[0], days[1], days[2], days[3], days[4], days[5], days[6])

  return output
}
