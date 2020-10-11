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
  var beginTime time.Time
  var finishTime time.Time
  var err error

  if begin == "" {
    beginTime = time.Now()
  } else {
    beginTime, err = ParseTime(begin)
    if err != nil {
      return Entry{}, err
    }
  }

  if finish != "" {
    finishTime, err = ParseTime(finish)
    if err != nil {
      return Entry{}, err
    }
  }

  return Entry{
    id,
    beginTime,
    finishTime,
    project,
    task,
    user,
  }, nil
}
