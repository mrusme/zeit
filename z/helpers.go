package z

import (
	"bytes"
	"errors"
	"math"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	TFAbsTwelveHour     int = 0
	TFAbsTwentyfourHour int = 1
	TFRelHourMinute     int = 2
	TFRelHourFraction   int = 3
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
func RelToTime(timeStr string, ftId int) (time.Time, error) {
	var re = regexp.MustCompile(TimeFormats()[ftId])
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

	switch gm[1] {
	case "+":
		t = time.Now().Local().Add(time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes))
	case "-":
		t = time.Now().Local().Add((time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes)) * -1)
	}

	return t, nil
}

func ParseTime(timeStr string) (time.Time, error) {
	tfId := GetTimeFormat(timeStr)

	t := time.Now()

	switch tfId {
	case TFAbsTwelveHour:
		tadj, err := time.Parse("3:04pm", timeStr)
		tnew := time.Date(t.Year(), t.Month(), t.Day(), tadj.Hour(), tadj.Minute(), t.Second(), t.Nanosecond(), t.Location())
		return tnew, err
	case TFAbsTwentyfourHour:
		tadj, err := time.Parse("15:04", timeStr)
		tnew := time.Date(t.Year(), t.Month(), t.Day(), tadj.Hour(), tadj.Minute(), t.Second(), t.Nanosecond(), t.Location())
		return tnew, err
	case TFRelHourMinute, TFRelHourFraction:
		return RelToTime(timeStr, tfId)
	default:
		return time.Now(), errors.New("could not match passed time")
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
	var _, cw = date.ISOWeek()
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
