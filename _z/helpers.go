package z

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/markusmobius/go-dateparser"
	"github.com/spf13/viper"
)

func TimeFormats() []string {
	return []string{
		`^\d{1,2}:\d{1,2}(am|pm)$`,     // Absolute twelve hour format
		`^\d{1,2}:\d{1,2}$`,            // Absolute twenty four hour format
		`^([+-])(\d{1,2}):(\d{1,2})$`,  // Relative hour:minute format
		`^([+-])(\d{1,2})\.(\d{1,2})$`, // Relative hour.fraction format
	}
}

func GetCurrentUser() string {
	user, err := user.Current()
	if err != nil {
		return "unknown"
	}

	return user.Username
}

func GetTimeFormat(timeStr string) int {
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

// TODO: Use https://golang.org/pkg/time/#ParseDuration
func RelToTime(timeStr string, ftId int, contextTime time.Time) (time.Time, error) {
	re := regexp.MustCompile(TimeFormats()[ftId])
	gm := re.FindStringSubmatch(timeStr)

	if len(gm) < 4 {
		return time.Now(), errors.New("No match")
	}

	var hours int = 0
	var minutes int = 0

	if ftId == TFRelHourFraction {
		f, _ := strconv.ParseFloat(gm[2]+"."+gm[3], 32)
		minutes = int(f * 60.0)
	} else {
		hours, _ = strconv.Atoi(gm[2])
		minutes, _ = strconv.Atoi(gm[3])
	}

	var t time.Time

	if viper.IsSet("time.relative") && viper.GetString("time.relative") == "context" && !contextTime.IsZero() {
		switch gm[1] {
		case "+":
			t = contextTime.Add(time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes))
		case "-":
			t = contextTime.Add((time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes)) * -1)
		}

		return t, nil
	}

	switch gm[1] {
	case "+":
		t = time.Now().Local().Add(time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes))
	case "-":
		t = time.Now().Local().Add((time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes)) * -1)
	}

	return t, nil
}

func ParseTime(timeStr string, contextTime time.Time) (time.Time, error) {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return time.Now(), errors.New("could not load location")
	}

	cfg := dateparser.Configuration{
		DefaultTimezone: loc,
	}

	tfId := GetTimeFormat(timeStr)

	switch tfId {
	case TFRelHourMinute, TFRelHourFraction:
		return RelToTime(timeStr, tfId, contextTime)
	default:
		tnew, err := dateparser.Parse(&cfg, timeStr)
		if err != nil {
			return time.Now(), errors.New("could not match passed time")
		}

		return tnew.Time, err
	}
}

func GetIdFromName(name string) string {
	reg, regerr := regexp.Compile("[^a-zA-Z0-9]+")
	if regerr != nil {
		return ""
	}

	id := strings.ToLower(reg.ReplaceAllString(name, ""))

	return id
}

func GetISOCalendarWeek(date time.Time) int {
	_, cw := date.ISOWeek()
	return cw
}

func GetISOWeekInMonth(date time.Time) (month int, weeknumber int) {
	if date.IsZero() {
		return -1, -1
	}

	newDay := (date.Day() - int(date.Weekday()) + 1)
	addDay := (date.Day() - newDay) * -1
	changedDate := date.AddDate(0, 0, addDay)

	return int(changedDate.Month()), int(math.Ceil(float64(changedDate.Day()) / 7.0))
}

func GetGitLog(repo string, since time.Time, until time.Time) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "-C", repo, "config", "user.name")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", "", err
	}
	gitUserStr, gitUserErrStr := string(stdout.Bytes()), string(stderr.Bytes())
	if gitUserStr == "" && gitUserErrStr != "" {
		return gitUserStr, gitUserErrStr, errors.New(gitUserErrStr)
	}

	stdout.Reset()
	stderr.Reset()

	cmd = exec.Command("git", "-C", repo, "log", "--author", gitUserStr, "--since", since.Format("2006-01-02T15:04:05-0700"), "--until", until.Format("2006-01-02T15:04:05-0700"), "--pretty=oneline")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return "", "", err
	}

	stdoutStr, stderrStr := string(stdout.Bytes()), string(stderr.Bytes())
	return stdoutStr, stderrStr, nil
}

func Ranges() []string {
	return []string{
		"today",
		"yesterday",
		"thisWeek",
		"lastWeek",
		"thisMonth",
		"lastMonth",
	}
}

func ParseSinceUntil(since string, until string, listRange string) (time.Time, time.Time) {
	var sinceTime time.Time
	var untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = now.Parse(since)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}
	}

	if until != "" {
		untilTime, err = now.Parse(until)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}
	}

	if listRange != "" {
		if since != "" || until != "" {
			fmt.Println("Range and since/until can't be used together, select one of them")
			os.Exit(1)
		}

		if viper.GetBool("firstWeekDayMonday") {
			now.WeekStartDay = time.Monday
		}

		loc, _ := time.LoadLocation("Local")
		time.Local = loc
		switch strings.ToLower(listRange) {
		case "today":
			sinceTime = now.BeginningOfDay()
			untilTime = now.EndOfDay()
		case "yesterday":
			sinceTime = now.BeginningOfDay().AddDate(0, 0, -1)
			untilTime = now.EndOfDay().AddDate(0, 0, -1)
		case "thisweek":
			sinceTime = now.BeginningOfWeek()
			untilTime = now.EndOfWeek()
		case "lastweek":
			lastWeekDay := time.Now().AddDate(0, 0, -7)
			sinceTime = now.With(lastWeekDay).BeginningOfWeek()
			untilTime = now.With(lastWeekDay).EndOfWeek()
		case "thismonth":
			sinceTime = now.BeginningOfMonth()
			untilTime = now.EndOfMonth()
		case "lastmonth":
			lastMonthDay := time.Now().AddDate(0, -1, 0)
			sinceTime = now.With(lastMonthDay).BeginningOfMonth()
			untilTime = now.With(lastMonthDay).EndOfMonth()
		default:
			fmt.Println("Unknown range selection, possible options: ", strings.Join(Ranges(), " "))
			os.Exit(1)
		}
	}

	return sinceTime, untilTime
}
