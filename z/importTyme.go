package z

import (
  "encoding/json"
  // "fmt"
  "os"
  // "github.com/shopspring/decimal"
  // "time"
)

type TymeEntry struct {
  Billing string `json:"billing"` // "UNBILLED",
  Category string `json:"category"` // "Client",
  Distance string `json:"distance"` // "0",
  Duration string `json:"duration"` // "15",
  End string `json:"end"` // "2020-09-01T08:57:00+01:00",
  Note string `json:"note"` // "",
  Project string `json:"project"` // "Project",
  Quantity string `json:"quantity"` // "0",
  Rate string `json:"rate"` // "140",
  RoundingMethod string `json:"rounding_method"` // "NEAREST",
  RoundingMinutes int `json:"rounding_minutes"` // 15,
  Start string `json:"start"` // "2020-09-01T08:45:00+01:00",
  Subtask string `json:"subtask"` // "",
  Sum string `json:"sum"` // "35",
  Task string `json:"task"` // "Development",
  Type string `json:"type"` // "timed",
  User string `json:"user"` // ""
}

type Tyme struct {
  Data []TymeEntry `json:"data"`
}

func (tyme *Tyme) Load(filename string) error {
  file, err := os.Open(filename)
  if err != nil {
    return err
  }
  defer file.Close()

  decoder := json.NewDecoder(file)

  if err = decoder.Decode(&tyme); err != nil {
    return err
  }

  return nil
}
