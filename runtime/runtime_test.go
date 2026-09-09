package runtime

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/spf13/cobra"
	"xn--gckvb8fzb.com/zeit/helpers/log"
	"xn--gckvb8fzb.com/zeit/helpers/out"
)

func newTestCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "zeit"}

	cmd.PersistentFlags().Bool("debug", false, "")
	cmd.PersistentFlags().String("color", "auto", "")
	cmd.PersistentFlags().String("text", "", "")
	cmd.PersistentFlags().Int("number", 0, "")

	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags(%v) = %s", args, err)
	}

	return cmd
}

func TestGetLogLevel(t *testing.T) {
	if got := GetLogLevel(newTestCommand(t)); got != slog.LevelError {
		t.Errorf("GetLogLevel() = %s, want %s", got, slog.LevelError)
	}

	if got := GetLogLevel(
		newTestCommand(t, "--debug"),
	); got != slog.LevelDebug {
		t.Errorf("GetLogLevel() = %s, want %s", got, slog.LevelDebug)
	}
}

func TestGetLogLevelWithoutTheFlag(t *testing.T) {
	if got := GetLogLevel(&cobra.Command{Use: "bare"}); got != slog.LevelError {
		t.Errorf("GetLogLevel() = %s, want %s", got, slog.LevelError)
	}
}

func TestGetOutputColor(t *testing.T) {
	tests := []struct {
		value string
		want  out.OutputColor
	}{
		{value: "never", want: out.ColorNever},
		{value: "auto", want: out.ColorAuto},
		{value: "always", want: out.ColorAlways},
		{value: "ALWAYS", want: out.ColorAlways},
		{value: "Never", want: out.ColorNever},
		{value: "nonsense", want: out.ColorAuto},
		{value: "", want: out.ColorAuto},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			cmd := newTestCommand(t, "--color", test.value)

			if got := GetOutputColor(cmd); got != test.want {
				t.Errorf("GetOutputColor(%q) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}

func TestGetOutputColorWithoutTheFlag(t *testing.T) {
	if got := GetOutputColor(&cobra.Command{Use: "bare"}); got != out.ColorAuto {
		t.Errorf("GetOutputColor() = %d, want ColorAuto", got)
	}
}

func TestAliasMapGetAliases(t *testing.T) {
	amap := AliasMap{
		"start":  {"started", "sta", "s"},
		"switch": {"switched", "sw"},
	}

	got := amap.GetAliases()

	if len(got) != 5 {
		t.Fatalf("GetAliases() = %v, want five entries", got)
	}
	for _, alias := range []string{"started", "sta", "s", "switched", "sw"} {
		if slices.Contains(got, alias) == false {
			t.Errorf("GetAliases() = %v, missing %q", got, alias)
		}
	}
}

func TestAliasMapGetCommandNameForAlias(t *testing.T) {
	amap := AliasMap{
		"start":  {"started", "sta", "s"},
		"switch": {"switched", "switch", "sw"},
		"resume": {"resume", "re"},
	}

	tests := []struct {
		alias string
		want  string
	}{
		{alias: "start", want: "start"},
		{alias: "started", want: "start"},
		{alias: "s", want: "start"},
		{alias: "switch", want: "switch"},
		{alias: "sw", want: "switch"},
		{alias: "resume", want: "resume"},
		{alias: "re", want: "resume"},
		{alias: "unknown", want: ""},
		{alias: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.alias, func(t *testing.T) {
			if got := amap.GetCommandNameForAlias(test.alias); got != test.want {
				t.Errorf("GetCommandNameForAlias(%q) = %q, want %q",
					test.alias, got, test.want)
			}
		})
	}
}

func TestGetCommandCall(t *testing.T) {
	rt := new(Runtime)

	if got := rt.GetCommandCall(&cobra.Command{Use: "start [flags]"}); got != "start" {
		t.Errorf("GetCommandCall() = %q, want start", got)
	}
}

func TestGetDynamicSuggestions(t *testing.T) {
	rt := new(Runtime)
	possible := []string{"alpha", "beta", "alphabet", "gamma"}

	tests := []struct {
		prefix string
		want   []string
	}{
		{prefix: "", want: []string{"alpha", "beta", "alphabet", "gamma"}},
		{prefix: "a", want: []string{"alpha", "alphabet"}},
		{prefix: "alpha", want: []string{"alpha", "alphabet"}},
		{prefix: "alphab", want: []string{"alphabet"}},
		{prefix: "z", want: nil},
		{prefix: "alphabetical", want: nil},
	}

	for _, test := range tests {
		t.Run(test.prefix, func(t *testing.T) {
			got := rt.GetDynamicSuggestions(test.prefix, possible)

			if slices.Equal(got, test.want) == false {
				t.Errorf("GetDynamicSuggestions(%q) = %v, want %v",
					test.prefix, got, test.want)
			}
		})
	}
}

func TestGetDynamicSuggestionsOnAnEmptyList(t *testing.T) {
	rt := new(Runtime)

	if got := rt.GetDynamicSuggestions("a", nil); got != nil {
		t.Errorf("GetDynamicSuggestions() = %v, want nothing", got)
	}
}

func TestGetFlags(t *testing.T) {
	rt := &Runtime{Logger: log.New(slog.LevelError)}
	cmd := newTestCommand(t, "--text", "value", "--number", "7", "--debug")

	if got := rt.GetStringFlag(cmd, "text"); got != "value" {
		t.Errorf("GetStringFlag() = %q, want value", got)
	}
	if got := rt.GetIntFlag(cmd, "number"); got != 7 {
		t.Errorf("GetIntFlag() = %d, want 7", got)
	}
	if got := rt.GetBoolFlag(cmd, "debug"); got != true {
		t.Errorf("GetBoolFlag() = %t, want true", got)
	}
	if got := rt.GetDebugFlag(cmd); got != true {
		t.Errorf("GetDebugFlag() = %t, want true", got)
	}
	if got := rt.IsDebug(cmd); got != true {
		t.Errorf("IsDebug() = %t, want true", got)
	}
	if got := rt.GetColorFlag(cmd); got != "auto" {
		t.Errorf("GetColorFlag() = %q, want auto", got)
	}
}

func TestGetFlagsFallBackWhenTheFlagIsMissing(t *testing.T) {
	rt := &Runtime{Logger: log.New(slog.LevelError + 1)}
	cmd := &cobra.Command{Use: "bare"}

	if got := rt.GetStringFlag(cmd, "absent"); got != "" {
		t.Errorf("GetStringFlag() = %q, want an empty string", got)
	}
	if got := rt.GetIntFlag(cmd, "absent"); got != 0 {
		t.Errorf("GetIntFlag() = %d, want 0", got)
	}
	if got := rt.GetBoolFlag(cmd, "absent"); got != false {
		t.Errorf("GetBoolFlag() = %t, want false", got)
	}
}

func TestNewBuildFallsBackToBuildInfo(t *testing.T) {
	got := NewBuild()

	if got.Version == "" {
		t.Errorf("Version is empty, want a value from the build info")
	}
}

func TestNewBuildPrefersLinkerValues(t *testing.T) {
	previous := Version
	Version = "v1.2.3"
	t.Cleanup(func() { Version = previous })

	if got := NewBuild(); got.Version != "v1.2.3" {
		t.Errorf("Version = %q, want the linker value", got.Version)
	}
}

func TestNewPersistsTheUserKey(t *testing.T) {
	t.Setenv(DATABASE_ENV_VAR, t.TempDir())

	first := New(slog.LevelError, out.ColorNever, false)
	firstKey := first.GetUserKey()
	first.End()

	if firstKey == "" {
		t.Fatalf("GetUserKey() is empty")
	}

	second := New(slog.LevelError, out.ColorNever, false)
	secondKey := second.GetUserKey()
	second.End()

	if firstKey != secondKey {
		t.Errorf("user key changed between runs: %q then %q", firstKey, secondKey)
	}
}

func TestNewPopulatesTheRuntime(t *testing.T) {
	t.Setenv(DATABASE_ENV_VAR, t.TempDir())

	rt := New(slog.LevelError, out.ColorNever, false)
	defer rt.End()

	if rt.Logger == nil {
		t.Errorf("Logger is nil")
	}
	if rt.Out == nil {
		t.Errorf("Out is nil")
	}
	if rt.Database == nil {
		t.Errorf("Database is nil")
	}
	if rt.Config == nil {
		t.Errorf("Config is nil")
	}
	if rt.Build.Version == "" {
		t.Errorf("Build.Version is empty")
	}
}
