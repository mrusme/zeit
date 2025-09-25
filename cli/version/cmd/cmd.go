package versionCmd

import (
	"github.com/charmbracelet/lipgloss/v2"
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
		if rt.Out.InColor() {
			style := lipgloss.NewStyle().Foreground(lipgloss.BrightBlack)
			rt.Out.Put(out.Info,
				"%s %s\n  %s %s\n  %s %s\n",
				style.Render("zeit"),
				rt.Build.Version,
				style.Render("Commit:"),
				rt.Build.Commit,
				style.Render("Build date:"),
				rt.Build.Date,
			)
		} else {
			rt.Out.Put(out.Info,
				"%s %s\n  %s %s\n  %s %s\n",
				"zeit",
				rt.Build.Version,
				"Commit:",
				rt.Build.Commit,
				"Build date:",
				rt.Build.Date,
			)
		}
	},
}

func init() {
}
