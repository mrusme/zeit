package runtime

import (
	"log/slog"
	"strings"

	"github.com/mrusme/zeit/helpers/out"
	"github.com/spf13/cobra"
)

func GetLogLevel(cmd *cobra.Command) slog.Level {
	var lvl slog.Level = slog.LevelInfo
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
