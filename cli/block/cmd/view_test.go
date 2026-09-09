package blockCmd

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

func TestBlockViewCarriesEveryKey(t *testing.T) {
	got := decodeView(t, BlockView{})

	for _, key := range []string{
		"key", "project_sid", "task_sid", "note", "start", "end", "duration",
	} {
		if _, ok := got[key]; ok == false {
			t.Errorf("key %q is missing", key)
		}
	}
}

func TestBlockViewReportsAMissingEndAsNull(t *testing.T) {
	got := decodeView(t, BlockView{})

	if got["end"] != nil {
		t.Errorf("end = %v, want null", got["end"])
	}
}

func TestBlockViewReportsDurationsInSeconds(t *testing.T) {
	got := decodeView(t, BlockView{Duration: out.Seconds(2 * time.Hour)})

	if got["duration"] != float64(7200) {
		t.Errorf("duration = %v, want 7200", got["duration"])
	}
}

func TestBlockViewKeepsRealEndTimestamps(t *testing.T) {
	ts := time.Date(2026, time.August, 26, 17, 0, 0, 0, time.UTC)

	got := decodeView(t, BlockView{TimestampEnd: out.EndTimestamp(ts)})

	if got["end"] == nil {
		t.Errorf("end = null, want a timestamp")
	}
}
