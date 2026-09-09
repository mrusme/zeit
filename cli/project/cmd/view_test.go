package projectCmd

import (
	"encoding/json"
	"testing"
	"time"

	"xn--gckvb8fzb.com/zeit/helpers/out"
)

func decodeView(t *testing.T, v any) map[string]any {
	t.Helper()

	encoded, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() = %s", err)
	}

	decoded := make(map[string]any)
	if err = json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(%s) = %s", encoded, err)
	}

	return decoded
}

func TestProjectViewCarriesEveryKey(t *testing.T) {
	got := decodeView(t, ProjectView{})

	for _, key := range []string{
		"sid", "display_name", "color", "tasks", "total_blocks", "total_amount",
	} {
		if _, ok := got[key]; ok == false {
			t.Errorf("key %q is missing", key)
		}
	}
}

func TestProjectTaskViewCarriesEveryKey(t *testing.T) {
	got := decodeView(t, ProjectTaskView{})

	for _, key := range []string{
		"sid", "display_name", "color", "total_blocks", "total_amount",
	} {
		if _, ok := got[key]; ok == false {
			t.Errorf("key %q is missing", key)
		}
	}
}

func TestProjectViewReportsAmountsInSeconds(t *testing.T) {
	got := decodeView(t, ProjectView{TotalAmount: out.Seconds(3 * time.Hour)})

	if got["total_amount"] != float64(10800) {
		t.Errorf("total_amount = %v, want 10800", got["total_amount"])
	}
}
