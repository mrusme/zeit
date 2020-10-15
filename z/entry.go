package z

import (
  "errors"
  "time"
  "fmt"
  "github.com/gookit/color"
)

type Entry struct {
  ID      string      `json:"-"`
  Begin   time.Time   `json:"begin,omitempty"`
  Finish  time.Time   `json:"finish,omitempty"`
  Project string      `json:"project,omitempty"`
  Task    string      `json:"task,omitempty"`
  User    string      `json:"user,omitempty"`
}

func NewEntry(
  id string,
  begin string,
  finish string,
  project string,
  task string,
  user string) (Entry, error) {
  var err error

  newEntry := Entry{}

  newEntry.ID = id
  newEntry.Project = project
  newEntry.Task = task
  newEntry.User = user

  _, err = newEntry.SetBeginFromString(begin)
  if err != nil {
    return Entry{}, err
  }

  _, err = newEntry.SetFinishFromString(finish)
  if err != nil {
    return Entry{}, err
  }

  if newEntry.IsFinishedAfterBegan() == false {
    return Entry{}, errors.New("beginning time of tracking cannot be after finish time")
  }

  return newEntry, nil
}

func (entry *Entry) SetBeginFromString(begin string) (time.Time, error) {
  var beginTime time.Time
  var err error

  if begin == "" {
    beginTime = time.Now()
  } else {
    beginTime, err = ParseTime(begin)
    if err != nil {
      return beginTime, err
    }
  }

  entry.Begin = beginTime
  return beginTime, nil
}

func (entry *Entry) SetFinishFromString(finish string) (time.Time, error) {
  var finishTime time.Time
  var err error

  if finish != "" {
    finishTime, err = ParseTime(finish)
    if err != nil {
      return finishTime, err
    }
  }

  entry.Finish = finishTime
  return finishTime, nil
}

func (entry *Entry) IsFinishedAfterBegan() (bool) {
  return (entry.Finish.IsZero() || entry.Begin.Before(entry.Finish))
}

func (entry *Entry) GetOutputForTrack(isRunning bool) (string) {
  outputPrefix := "began tracking"
  if isRunning == false {
    outputPrefix = "tracked"
  }

  if entry.Task != "" && entry.Project != "" {
    return fmt.Sprintf("▷ %s %s on %s\n", outputPrefix, color.FgLightWhite.Render(entry.Task), color.FgLightWhite.Render(entry.Project))
  } else if entry.Task != "" && entry.Project == "" {
    return fmt.Sprintf("▷ %s %s\n", outputPrefix, color.FgLightWhite.Render(entry.Task))
  } else if entry.Task == "" && entry.Project != "" {
    return fmt.Sprintf("▷ %s task on %s\n", outputPrefix, color.FgLightWhite.Render(entry.Project))
  } else {
    return fmt.Sprintf("▷ %s task\n", outputPrefix)
  }
}

func (entry *Entry) GetOutputForFinish() (string) {
  trackDiff := entry.Finish.Sub(entry.Begin)
  trackDiffOut := time.Time{}.Add(trackDiff)

  if entry.Task != "" && entry.Project != "" {
    return fmt.Sprintf("□ finished tracking %s on %s for %sh\n", color.FgLightWhite.Render(entry.Task), color.FgLightWhite.Render(entry.Project), trackDiffOut.Format("15:04"))
  } else if entry.Task != "" && entry.Project == "" {
    return fmt.Sprintf("□ finished tracking %s for %sh\n", color.FgLightWhite.Render(entry.Task), trackDiffOut.Format("15:04"))
  } else if entry.Task == "" && entry.Project != "" {
    return fmt.Sprintf("□ finished tracking task on %s for %sh\n", color.FgLightWhite.Render(entry.Project), trackDiffOut.Format("15:04"))
  }

  return fmt.Sprintf("□ finished tracking task\n")
}
