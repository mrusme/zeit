package z

import (
)

type Task struct {
  Name          string      `json:"name,omitempty"`
  GitRepository string      `json:"gitRepository,omitempty"`
}
