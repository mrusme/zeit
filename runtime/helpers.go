package runtime

import (
	"log/slog"

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
