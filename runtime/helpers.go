package runtime

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/mrusme/zeit/helpers/out"
	"github.com/spf13/cobra"
)

type AliasMap map[string][]string

func (amap *AliasMap) GetAliases() []string {
	var a []string
	for _, aliases := range *amap {
		a = append(a, aliases...)
	}
	return a
}

func (amap *AliasMap) GetCommandNameForAlias(alias string) string {
	for k, aliases := range *amap {
		if k == alias {
			return k
		}
		if slices.Contains(aliases, alias) == true {
			return k
		}
	}

	return ""
}

func GetLogLevel(cmd *cobra.Command) slog.Level {
	var lvl slog.Level = slog.LevelError
	flagDebug, _ := cmd.Flags().GetBool("debug")
	if flagDebug {
		lvl = slog.LevelDebug
	}

	return lvl
}

func GetOutputColor(cmd *cobra.Command) out.OutputColor {
	var oc out.OutputColor = out.ColorAlways
	flagColor, _ := cmd.Flags().GetString("color")
	switch strings.ToLower(flagColor) {
	case "never":
		oc = out.ColorNever
	case "auto":
		oc = out.ColorAuto
	case "always":
		oc = out.ColorAlways
	}

	return oc
}
