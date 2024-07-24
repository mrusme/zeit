package z

import (
	"encoding/json"
	// "fmt"
	"os"
	"time"

	"github.com/shopspring/decimal"
)

type TymeEntry struct {
	Billing         string `json:"billing"`          // "UNBILLED",
	Category        string `json:"category"`         // "Client",
	Distance        string `json:"distance"`         // "0",
	Duration        string `json:"duration"`         // "15",
	Start           string `json:"start"`            // "2020-09-01T08:45:00+01:00",
	End             string `json:"end"`              // "2020-09-01T08:57:00+01:00",
	Hours           string `json:"hours"`            // "3.35"
	Date            string `json:"date"`             // "30-12-2023" (DD-MM-YYYY)
	Note            string `json:"note"`             // "",
	Project         string `json:"project"`          // "Project",
	Quantity        string `json:"quantity"`         // "0",
	Rate            string `json:"rate"`             // "140",
	RoundingMethod  string `json:"rounding_method"`  // "NEAREST",
	RoundingMinutes int    `json:"rounding_minutes"` // 15,
	Subtask         string `json:"subtask"`          // "",
	Sum             string `json:"sum"`              // "35",
	Task            string `json:"task"`             // "Development",
	Type            string `json:"type"`             // "timed",
	User            string `json:"user"`             // ""
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

func (tyme *Tyme) FromEntries(entries []Entry) error {
	for _, entry := range entries {
		duration := decimal.NewFromFloat(entry.Finish.Sub(entry.Begin).Minutes())

		tymeEntry := TymeEntry{
			Billing:         "UNBILLED",
			Category:        "",
			Distance:        "0",
			Duration:        duration.StringFixed(0),
			Start:           entry.Begin.Format(time.RFC3339),
			End:             entry.Finish.Format(time.RFC3339),
			Hours:           entry.Hours,
			Date:            entry.Date,
			Note:            entry.Notes,
			Project:         entry.Project,
			Quantity:        "0",
			Rate:            "0",
			RoundingMethod:  "NEAREST",
			RoundingMinutes: 15,
			Subtask:         "",
			Sum:             "0",
			Task:            entry.Task,
			Type:            "timed",
			User:            "",
		}

		tyme.Data = append(tyme.Data, tymeEntry)
	}

	return nil
}

func (tyme *Tyme) Stringify() string {
	stringified, err := json.Marshal(tyme)
	if err != nil {
		return ""
	}

	return string(stringified)
}
