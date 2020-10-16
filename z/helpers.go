package z

import (
  "fmt"
  "os/user"
  "regexp"
  "strconv"
  "time"
  "errors"
)


const (
  TFAbsTwelveHour int = 0
  TFAbsTwentyfourHour int = 1
  TFRelHourMinute int = 2
  TFRelHourFraction int = 3
)

func TimeFormats() []string {
  return []string{
    `^\d{1,2}:\d{1,2}(am|pm)$`, // Absolute twelve hour format
    `^\d{1,2}:\d{1,2}$`, // Absolute twenty four hour format
    `^([+-])(\d{1,2}):(\d{1,2})$`, // Relative hour:minute format
    `^([+-])(\d{1,2})\.(\d{1,2})$`, // Relative hour.fraction format
  }
}

func GetCurrentUser() (string) {
  user, err := user.Current()
  if err != nil {
    return "unknown"
  }

  return user.Username
}

func GetTimeFormat(timeStr string) (int) {
  var matched bool
  var regerr error

  for timeFormatId, timeFormat := range TimeFormats() {
    matched, regerr = regexp.MatchString(timeFormat, timeStr)
    if regerr != nil {
      return -1
    }

    if matched == true {
      return timeFormatId
    }
  }

  return -1
}

func RelToTime(timeStr string, ftId int) (time.Time, error) {
    var re = regexp.MustCompile(TimeFormats()[ftId])
    gm := re.FindStringSubmatch(timeStr)

    if len(gm) < 4 {
      return time.Now(), errors.New("No match")
    }


    var hours int = 0
    var minutes int = 0

    if ftId == TFRelHourFraction {
      f, _ := strconv.ParseFloat(gm[2] + "." + gm[3], 32)
      minutes = int(f * 60.0)
    } else {
      hours, _ = strconv.Atoi(gm[2])
      minutes, _ = strconv.Atoi(gm[3])
    }

    var t time.Time

    switch gm[1] {
    case "+":
      t = time.Now().Local().Add(time.Hour * time.Duration(hours) + time.Minute * time.Duration(minutes))
    case "-":
      t = time.Now().Local().Add((time.Hour * time.Duration(hours) + time.Minute * time.Duration(minutes)) * -1)
    }

    return t, nil
}

func ParseTime(timeStr string) (time.Time, error) {
  tfId := GetTimeFormat(timeStr)

  switch tfId {
  case TFAbsTwelveHour:
    return time.Parse("3:04pm", timeStr)
  case TFAbsTwentyfourHour:
    return time.Parse("15:04", timeStr)
  case TFRelHourMinute, TFRelHourFraction:
    return RelToTime(timeStr, tfId)
  default:
    return time.Now(), errors.New("No match")
  }
}

func OutputAppendRight(leftStr string, rightStr string, pad int) (string) {
  var output string = ""
  var rpos int = 0

  left := []rune(leftStr)
  leftLen := len(left)
  right := []rune(rightStr)
  rightLen := len(right)

  for lpos := 0; lpos < leftLen; lpos++ {
    if left[lpos] == '\n' || lpos == (leftLen - 1) {
      output = fmt.Sprintf("%s%*s", output, pad, "")
      for rpos = rpos; rpos < rightLen; rpos++ {
        output = fmt.Sprintf("%s%c", output, right[rpos])
        if right[rpos] == '\n' {
          rpos++
          break
        }
      }
      continue
    }
    output = fmt.Sprintf("%s%c", output, left[lpos])
  }

  return output
}
