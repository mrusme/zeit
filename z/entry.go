package z

import (
  "errors"
  "time"
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
  return (entry.Finish.IsZero() == false && entry.Begin.Before(entry.Finish))
}
