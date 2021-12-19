package z

import (
  "fmt"
  "time"
  "strings"
  "github.com/shopspring/decimal"
  "github.com/jinzhu/now"
  // "github.com/gookit/color"
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
  Distribution map[string]Statistic
  TotalHours decimal.Decimal
}

func NewCalendar(entries []Entry) (Calendar, error) {
  cal := Calendar{}

  cal.Distribution = make(map[string]Statistic)

  projects := make(map[string]Project)

  for _, entry := range entries {
    var entryFinish time.Time
    endOfBeginDay := now.With(entry.Begin).EndOfDay()
    sameDayHours := decimal.NewFromInt(0)
    nextDayHours := decimal.NewFromInt(0)

    projectId := GetIdFromName(entry.Project)

    if projects[projectId].Name == "" {
      project, err := database.GetProject(entry.User, entry.Project)
      if err != nil {
        return cal, err
      }

      projects[projectId] = project
    }

    if entry.Finish.IsZero() {
      entryFinish = time.Now()
    } else {
      entryFinish = entry.Finish
    }

    /*
     * Apparently the activity end is on a new day.
     * This means we have to split the activity across two days.
     */
    if endOfBeginDay.Before(entryFinish) == true {
      startOfFinishDay := now.With(entryFinish).BeginningOfDay()

      sameDayDuration := endOfBeginDay.Sub(entry.Begin)
      sameDay := sameDayDuration.Hours()
      sameDayHours = decimal.NewFromFloat(sameDay)

      nextDayDuration := entryFinish.Sub(startOfFinishDay)
      nextDay := nextDayDuration.Hours()
      nextDayHours = decimal.NewFromFloat(nextDay)

    } else {
      sameDayDuration := entryFinish.Sub(entry.Begin)
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
        Color: GetColorFnFromHex(projects[projectId].Color),
      }

      if cal.Months[month0].Weeks[weeknumber0].Statistics == nil {
        cal.Months[month0].Weeks[weeknumber0].Statistics = make(WeekStatistics)
      }

      cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName] = append(cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName], stat)
    }

    if nextDayHours.GreaterThan(decimal.NewFromInt(0)) {
      month, weeknumber := GetISOWeekInMonth(entryFinish)
      month0 := month - 1
      weeknumber0 := weeknumber - 1
      weekday := entry.Begin.Weekday()
      weekdayName := weekday.String()[:2]

      stat := Statistic{
        Hours: nextDayHours,
        Project: entry.Project,
        Color: GetColorFnFromHex(projects[projectId].Color),
      }

      if cal.Months[month0].Weeks[weeknumber0].Statistics == nil {
        cal.Months[month0].Weeks[weeknumber0].Statistics = make(WeekStatistics)
      }

      cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName] = append(cal.Months[month0].Weeks[weeknumber0].Statistics[weekdayName], stat)
    }

    var dist = cal.Distribution[entry.Project]
    dist.Project = entry.Project
    dist.Hours = dist.Hours.Add(sameDayHours)
    dist.Hours = dist.Hours.Add(nextDayHours)
    dist.Color = GetColorFnFromHex(projects[projectId].Color)
    cal.Distribution[entry.Project] = dist

    // fmt.Printf("Same Day: %s \n Next Day: %s \n Project Hours: %s\n", sameDayHours.String(), nextDayHours.String(), dist.Hours.String())
    cal.TotalHours = cal.TotalHours.Add(sameDayHours)
    cal.TotalHours = cal.TotalHours.Add(nextDayHours)
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

    if dayHours.GreaterThan(decimal.NewFromInt(24)) {
      fmt.Printf("%s %s of week %d in month %d has more than 24h tracked; cutting at 24h now\n", CharError, day, (month+1), (week+1))
      dayHours = decimal.NewFromInt(24)
    }

    bar := GetOutputBarForHours(dayHours, calendar.Months[month].Weeks[week].Statistics[day])
    bars = append(bars, bar)
  }

  output = fmt.Sprintf("CW %02d                    %s H\n", GetISOCalendarWeek(date), fmtHours(totalHours))
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

func (calendar *Calendar) GetOutputForDistribution() (string) {
  var output string = ""

  // fmt.Printf("%s\n", calendar.TotalHours.String())

  var bar string = ""
  for _, stat := range calendar.Distribution {
    divided := stat.Hours.Div(calendar.TotalHours)
    percentage := divided.Mul(decimal.NewFromInt(100))
    hoursStr := fmtHours(stat.Hours)
    percentageStr := percentage.StringFixed(2)

    dividedByBarLength := percentage.Div(decimal.NewFromInt(100))
    percentageForBar := dividedByBarLength.Mul(decimal.NewFromInt(80))
    percentageForBarInt := int(percentageForBar.Round(0).IntPart())

    bar = fmt.Sprintf("%s%s", bar, stat.Color(strings.Repeat("█", percentageForBarInt)))

    output = fmt.Sprintf("%s%s%*s H / %*s %%\n", output, stat.Color(stat.Project), (68 - len(stat.Project)), hoursStr, 5, percentageStr)
  }

  output = fmt.Sprintf("DISTRIBUTION\n\n%s\n\n%s\n", bar, output)
  return output
}
