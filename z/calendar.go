package z

import (
  "fmt"
  "time"
  "github.com/shopspring/decimal"
)

type Calendar struct {

}

func GetOutputBoxForNumber(number int) (string) {
  switch(number) {
  case 0: return "  "
  case 1: return " ▄"
  case 2: return "▄▄"
  case 3: return "▄█"
  case 4: return "██"
  }

  return "  "
}

func GetOutputBarForHours(hours decimal.Decimal) ([]string) {
  var bar = []string{
    "····",
    "····",
    "····",
    "····",
    "····",
    "····",
  }

  hoursInt := int((hours.Round(0)).IntPart())
  rest := ((hours.Round(0)).Mod(decimal.NewFromInt(4))).Round(0)
  restInt := int(rest.IntPart())

  divisible := hoursInt - restInt
  fullparts := divisible / 4

  for i := (len(bar) - 1); i > (len(bar) - 1 - fullparts); i-- {
    bar[i] = " " + GetOutputBoxForNumber(4) + " "
  }

  if(restInt > 0) {
    bar[(len(bar) - 1 - fullparts)] = " " + GetOutputBoxForNumber(restInt) + " "
  }

  return bar
}

func (calendar *Calendar) GetCalendarWeek(timestamp time.Time) (int) {
  var _, cw = timestamp.ISOWeek()
  return cw
}

// func (calendar *Calendar) GetBufferForWeekCalendar(cw int, data map[string]decimal.Decimal) ([][]string) {
//   var output string = ""
//   var bars [][]string
//   var totalHours = decimal.NewFromInt(0)

//   var days = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
//   for _, day := range days {
//     hours := data[day]
//     totalHours = totalHours.Add(hours)
//     bar := GetOutputBarForHours(hours)
//     bars = append(bars, bar)
//   }

//   output = fmt.Sprintf("CW %02d                    %s H\n", cw, totalHours.StringFixed(2))
//   for row := 0; row < len(bars[0]); row++ {
//     output = fmt.Sprintf("%s%2d │", output, ((6 - row) * 4))
//     for col := 0; col < len(bars); col++ {
//       output = fmt.Sprintf("%s%s", output, bars[col][row])
//     }
//     output = fmt.Sprintf("%s\n", output)
//   }
//   output = fmt.Sprintf("%s   └────────────────────────────\n     %s  %s  %s  %s  %s  %s  %s",
//     output, days[0], days[1], days[2], days[3], days[4], days[5], days[6])

//   return output
// }

func (calendar *Calendar) GetTuiBufferForWeekCalendar(cw int, data map[string]decimal.Decimal) (TuiBuffer) {
  var output string = ""
  buffer := TuiBuffer{}
  var bars [][]string
  var totalHours = decimal.NewFromInt(0)

  var days = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
  for _, day := range days {
    hours := data[day]
    totalHours = totalHours.Add(hours)
    bar := GetOutputBarForHours(hours)
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
  output = fmt.Sprintf("%s   └────────────────────────────\n     %s  %s  %s  %s  %s  %s  %s",
    output, days[0], days[1], days[2], days[3], days[4], days[5], days[6])

  fmt.Printf("%s\n", output)

  row := 0
  col := 0
  for _, chr := range output {
    if(chr == '\n') {
      row++
      col = 0
      continue
    }

    buffer[row][col] = chr
    col++
  }
  return buffer
}
