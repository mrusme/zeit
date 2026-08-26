package out

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"charm.land/lipgloss/v2"
)

type EndTimestamp time.Time

func (ts EndTimestamp) MarshalJSON() ([]byte, error) {
	if time.Time(ts).IsZero() == true {
		return []byte("null"), nil
	}

	return json.Marshal(time.Time(ts))
}

func (ts EndTimestamp) String() string {
	if time.Time(ts).IsZero() == true {
		return "not ended"
	}

	return time.Time(ts).Format(time.DateTime)
}

type Seconds time.Duration

func (s Seconds) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(time.Duration(s).Seconds()))
}

func (s Seconds) String() string {
	return time.Duration(s).Round(time.Second).String()
}

type StatusOut struct {
	Status     string `json:"status"`
	IsRunning  bool   `json:"is_running"`
	ProjectSID string `json:"project_sid"`
	TaskSID    string `json:"task_sid"`
	Timer      int64  `json:"timer"`
}

func RandomVisibleHexColor() string {
	// Randomize R, G, B values within a mid-range (64 and 191) for better
	// contrast on light and dark backgrounds
	r := rand.Intn(128) + 64
	g := rand.Intn(128) + 64
	b := rand.Intn(128) + 64

	// Format the RGB values into hex color code
	hexColor := fmt.Sprintf("#%02X%02X%02X", r, g, b)
	return hexColor
}

func Color(cstr string) color.Color {
	return lipgloss.Color(cstr)
}
