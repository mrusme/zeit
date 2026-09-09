package importer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"xn--gckvb8fzb.com/zeit/database"
	"xn--gckvb8fzb.com/zeit/errs"
	"xn--gckvb8fzb.com/zeit/models/activeblock"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/models/config"
	"xn--gckvb8fzb.com/zeit/models/project"
	"xn--gckvb8fzb.com/zeit/models/task"
)

type collected struct {
	models []database.Model
	errors []error
}

func (c *collected) callback(entry database.Model, err error, v ...any) error {
	if err != nil {
		c.errors = append(c.errors, err)
		return nil
	}

	c.models = append(c.models, entry)

	return nil
}

func (c *collected) counts() map[string]int {
	counts := make(map[string]int)

	for _, model := range c.models {
		switch model.(type) {
		case *activeblock.ActiveBlock:
			counts["activeblock"]++
		case *block.Block:
			counts["block"]++
		case *config.Config:
			counts["config"]++
		case *project.Project:
			counts["project"]++
		case *task.Task:
			counts["task"]++
		}
	}

	return counts
}

func writeFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "import.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() = %s", err)
	}

	return path
}

const v1Payload = `{
  "config": {"user_key": "01a03add-f355-7d56-81a5-c30c895ba495"},
  "activeblock": {"active_block_key": "block:a", "previous_block_key": "block:b"},
  "project:01998b32-7f89-7373-a192-000000000001": {
    "sid": "alpha", "display_name": "Alpha", "color": "#112233"
  },
  "task:01998b32-7f89-7373-a192-000000000002": {
    "sid": "one", "project_sid": "alpha", "display_name": "One", "color": "#445566"
  },
  "block:01998b32-7f89-7373-a192-000000000003": {
    "project_sid": "alpha", "task_sid": "one", "note": "a note",
    "start": "2026-08-26T09:00:00Z", "end": "2026-08-26T10:00:00Z"
  }
}`

const v0Payload = `[
  {"begin":"2026-08-26T09:00:00Z","finish":"2026-08-26T10:00:00Z",
   "project":"Old Project","task":"Some Task","user":"m"},
  {"begin":"2026-08-27T09:00:00Z","finish":"2026-08-27T10:00:00Z",
   "project":"Old Project","task":"Other Task","user":"m"}
]`

func TestNewRejectsUnknownFormats(t *testing.T) {
	path := writeFile(t, v1Payload)

	for _, format := range []ImportFileType{"", "v2", "tyme", "V1"} {
		t.Run(string(format), func(t *testing.T) {
			im, err := New(format, path)

			if errors.Is(err, errs.ErrUnknownImportFormat) == false {
				t.Errorf("New(%q) = %v, want ErrUnknownImportFormat", format, err)
			}
			if im != nil {
				t.Errorf("New(%q) returned an importer alongside the error", format)
			}
		})
	}
}

func TestNewRejectsAMissingFile(t *testing.T) {
	if _, err := New(TypeZeitV1, filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Errorf("New() = nil, want an error for a missing file")
	}
}

func TestImportV1(t *testing.T) {
	im, err := New(TypeZeitV1, writeFile(t, v1Payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	if len(got.errors) != 0 {
		t.Errorf("Import() reported %d errors, want none", len(got.errors))
	}

	want := map[string]int{
		"activeblock": 1,
		"block":       1,
		"config":      1,
		"project":     1,
		"task":        1,
	}
	counts := got.counts()

	for name, count := range want {
		if counts[name] != count {
			t.Errorf("imported %d %s entries, want %d", counts[name], name, count)
		}
	}
}

func TestImportV1SetsKeys(t *testing.T) {
	im, err := New(TypeZeitV1, writeFile(t, v1Payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	for _, model := range got.models {
		if model.GetKey() == "" {
			t.Errorf("%T was imported without a key", model)
		}
	}
}

func TestImportV1SkipsUnknownModels(t *testing.T) {
	im, err := New(TypeZeitV1, writeFile(t, `{"mystery:1": {"field": true}}`))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	if len(got.models) != 0 {
		t.Errorf("Import() produced %d models, want none", len(got.models))
	}
	if len(got.errors) != 0 {
		t.Errorf("Import() reported %d errors, want none", len(got.errors))
	}
}

func TestImportV1SkipsCorruptEntries(t *testing.T) {
	payload := `{
  "project:01998b32-7f89-7373-a192-000000000001": {"sid": "alpha", "color": "#112233"},
  "project:01998b32-7f89-7373-a192-000000000002": "not an object",
  "task:01998b32-7f89-7373-a192-000000000003": {"sid": 12345}
}`

	im, err := New(TypeZeitV1, writeFile(t, payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	if len(got.models) != 1 {
		t.Errorf("Import() produced %d models, want only the valid one", len(got.models))
	}
	if len(got.errors) != 2 {
		t.Errorf("Import() reported %d errors, want 2", len(got.errors))
	}
	for _, err := range got.errors {
		if errors.Is(err, errs.ErrDataConversion) == false {
			t.Errorf("Import() reported %v, want ErrDataConversion", err)
		}
	}
}

func TestImportV1RejectsMalformedJSON(t *testing.T) {
	payloads := []string{
		"{not json at all",
		`{"project:1": }`,
		`{"project:1": {"sid": "alpha"`,
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			im, err := New(TypeZeitV1, writeFile(t, payload))
			if err != nil {
				t.Fatalf("New() = %s", err)
			}
			defer im.End()

			got := new(collected)
			if err = im.Import(got.callback); err == nil {
				t.Errorf("Import() = nil, want an error")
			}
		})
	}
}

func TestImportV1OnAnEmptyFile(t *testing.T) {
	im, err := New(TypeZeitV1, writeFile(t, ""))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	if len(got.models) != 0 {
		t.Errorf("Import() produced %d models, want none", len(got.models))
	}
}

func TestImportV1StopsWhenTheCallbackFails(t *testing.T) {
	im, err := New(TypeZeitV1, writeFile(t, v1Payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	sentinel := errors.New("stop here")
	calls := 0

	err = im.Import(func(entry database.Model, err error, v ...any) error {
		calls++
		return sentinel
	})

	if errors.Is(err, sentinel) == false {
		t.Errorf("Import() = %v, want the callback error", err)
	}
	if calls != 1 {
		t.Errorf("the callback was invoked %d times, want 1", calls)
	}
}

func TestImportV1PassesThroughExtraArguments(t *testing.T) {
	im, err := New(TypeZeitV1, writeFile(t, v1Payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	marker := "carried through"
	seen := 0

	err = im.Import(func(entry database.Model, err error, v ...any) error {
		if len(v) != 1 || v[0] != marker {
			t.Errorf("callback received %v, want %q", v, marker)
		}
		seen++
		return nil
	}, marker)
	if err != nil {
		t.Fatalf("Import() = %s", err)
	}

	if seen == 0 {
		t.Errorf("the callback was never invoked")
	}
}

func TestImportV0(t *testing.T) {
	im, err := New(TypeZeitV0, writeFile(t, v0Payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	counts := got.counts()

	if counts["block"] != 2 {
		t.Errorf("imported %d blocks, want 2", counts["block"])
	}
	if counts["project"] != 1 {
		t.Errorf("imported %d projects, want 1", counts["project"])
	}
	if counts["task"] != 2 {
		t.Errorf("imported %d tasks, want 2", counts["task"])
	}
}

func TestImportV0ConvertsNamesToSIDs(t *testing.T) {
	im, err := New(TypeZeitV0, writeFile(t, v0Payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	for _, model := range got.models {
		pj, ok := model.(*project.Project)
		if ok == false {
			continue
		}

		if pj.SID != "old_project" {
			t.Errorf("project SID = %q, want old_project", pj.SID)
		}
		if pj.DisplayName != "Old Project" {
			t.Errorf("project display name = %q, want Old Project", pj.DisplayName)
		}
	}
}

func TestImportV0CarriesTimestamps(t *testing.T) {
	im, err := New(TypeZeitV0, writeFile(t, v0Payload))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	for _, model := range got.models {
		b, ok := model.(*block.Block)
		if ok == false {
			continue
		}

		if b.TimestampStart.IsZero() == true || b.TimestampEnd.IsZero() == true {
			t.Errorf("block %q has no timestamps", b.GetKey())
		}
		if b.TimestampEnd.After(b.TimestampStart) == false {
			t.Errorf("block %q ends before it starts", b.GetKey())
		}
	}
}

func TestImportV0RejectsMalformedJSON(t *testing.T) {
	im, err := New(TypeZeitV0, writeFile(t, "[not json at all"))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err == nil {
		t.Errorf("Import() = nil, want an error")
	}
}

func TestImportV0OnAnEmptyList(t *testing.T) {
	im, err := New(TypeZeitV0, writeFile(t, "[]"))
	if err != nil {
		t.Fatalf("New() = %s", err)
	}
	defer im.End()

	got := new(collected)
	if err = im.Import(got.callback); err != nil {
		t.Fatalf("Import() = %s", err)
	}

	if len(got.models) != 0 {
		t.Errorf("Import() produced %d models, want none", len(got.models))
	}
}
