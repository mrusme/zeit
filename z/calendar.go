package z

import (
  "fmt"
  "time"
  "github.com/shopspring/decimal"
  "github.com/jinzhu/now"
  "github.com/gookit/color"
)

type Statistic struct {
  Hours decimal.Decimal
  Project string
  Color (func(...interface {}) string)
}

type WeekStatistics map[string][]Statistic

type Week struct {
  Statistics WeekStatistics
}

type Month struct {
  Name string
  Weeks [5]Week
}

type Calendar struct {
  Months [12]Month
}

func NewCalendar(entries []Entry) (Calendar, error) {
  cal := Calendar{}

  for _, entry := range entries {
    endOfBeginDay := now.With(entry.Begin).EndOfDay()
    sameDayHours := decimal.NewFromInt(0)
    nextDayHours := decimal.NewFromInt(0)

    /*
     * Apparently the activity end is on a new day.
     * This means we have to split the activity across two days.
     */
    if endOfBeginDay.Before(entry.Finish) == true {
      startOfFinishDay := now.With(entry.Finish).BeginningOfDay()

      sameDayDuration := endOfBeginDay.Sub(entry.Begin)
      sameDay := sameDayDuration.Hours()
      sameDayHours = decimal.NewFromFloat(sameDay)

      nextDayDuration := entry.Finish.Sub(startOfFinishDay)
      nextDay := nextDayDuration.Hours()
      nextDayHours = decimal.NewFromFloat(nextDay)

    } else {
      sameDayDuration := entry.Finish.Sub(entry.Begin)
      sameDay := sameDayDuration.Hours()
      sameDayHours = decimal.NewFromFloat(sameDay)
    }

    if sameDayHours.GreaterThan(decimal.NewFromInt(0)) {
      month, weeknumber := GetISOWeekInMonth(entry.Begin)
      month0 := month - 1
      weeknumber0 := weeknumber - 1
      weekday := entry.Begin.Weekday()
      weekdayName := weekday.String()[:2]

      stat := Statistic{
        Hours: sameDayHours,
        Project: entry.Project,
        Color: color.FgCyan.Render,
      }

      if cal.Months[month0].Weeks[weeknumber0].Statistics == nil {
        cal.Months[month0].Weeks[weeknumber0].Statistics = make(WeekStatistics)
      }

      cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName] = append(cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName], stat)
    }

    if nextDayHours.GreaterThan(decimal.NewFromInt(0)) {
      month, weeknumber := GetISOWeekInMonth(entry.Finish)
      month0 := month - 1
      weeknumber0 := weeknumber - 1
      weekday := entry.Begin.Weekday()
      weekdayName := weekday.String()[:2]

      stat := Statistic{
        Hours: nextDayHours,
        Project: entry.Project,
        Color: color.FgCyan.Render,
      }

      if cal.Months[month0].Weeks[weeknumber0].Statistics == nil {
        cal.Months[month0].Weeks[weeknumber0].Statistics = make(WeekStatistics)
      }

      cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName] = append(cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName], stat)
    }
  }

  return cal, nil
}

func (calendar *Calendar) GetOutputForWeekCalendar(date time.Time, month int, week int) (string) {
  var output string = ""
  var bars [][]string
  var totalHours = decimal.NewFromInt(0)

  var days = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
  for _, day := range days {
    var dayHours = decimal.NewFromInt(0)

    for _, stat := range calendar.Months[month].Weeks[week].Statistics[day] {
      dayHours = dayHours.Add(stat.Hours)
      totalHours = totalHours.Add(stat.Hours)
    }

    bar := GetOutputBarForHours(dayHours, calendar.Months[month].Weeks[week].Statistics[day])
    bars = append(bars, bar)
  }

  output = fmt.Sprintf("CW %02d                    %s H\n", GetISOCalendarWeek(date), totalHours.StringFixed(2))
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
