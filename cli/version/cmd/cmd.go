package versionCmd

import (
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var flagMachine string

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "zeit version",
	Long:  "Display zeit version information",
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()
		rt.Out.Put(out.Info,
			"%s %s\n  %s %s\n  %s %s\n",
			rt.Out.Stylize(
				out.Style{FG: out.ColorPrimary, BG: out.ColorSecondary},
				"zeit"),
			rt.Build.Version,
			rt.Out.FG(out.ColorSecondary, "Commit:"),
			rt.Build.Commit,
			rt.Out.FG(out.ColorSecondary, "Build date:"),
			rt.Build.Date,
		)
	},
}

func init() {
}
