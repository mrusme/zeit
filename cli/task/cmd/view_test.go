package taskCmd

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

func TestTaskViewCarriesEveryKey(t *testing.T) {
	got := decodeView(t, TaskView{})

	for _, key := range []string{
		"sid", "project_sid", "display_name", "color",
		"blocks", "total_blocks", "total_amount",
	} {
		if _, ok := got[key]; ok == false {
			t.Errorf("key %q is missing", key)
		}
	}
}

func TestTaskBlockViewCarriesEveryKey(t *testing.T) {
	got := decodeView(t, TaskBlockView{})

	for _, key := range []string{"key", "note", "start", "end", "duration"} {
		if _, ok := got[key]; ok == false {
			t.Errorf("key %q is missing", key)
		}
	}
}

func TestTaskViewReportsAmountsInSeconds(t *testing.T) {
	got := decodeView(t, TaskView{TotalAmount: out.Seconds(90 * time.Minute)})

	if got["total_amount"] != float64(5400) {
		t.Errorf("total_amount = %v, want 5400", got["total_amount"])
	}
}

func TestTaskBlockViewReportsAMissingEndAsNull(t *testing.T) {
	got := decodeView(t, TaskBlockView{})

	if got["end"] != nil {
		t.Errorf("end = %v, want null", got["end"])
	}
}
