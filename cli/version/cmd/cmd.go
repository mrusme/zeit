package versionCmd

import (
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "zeit version",
	Long:  "Display zeit version information",
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		if rt.Out.InColor() {
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                         ▊▊▊▊▊▊▊▊ ▊▊▊▊▊▊▊▊ ▊▊▊ ▊▊▊▊▊▊▊▊                         ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                               ▊▊ ▊         ▊  ▊▊                               ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                              ▊▊  ▊         ▊   ▊▊                              ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                             ▊▊   ▊         ▊    ▊▊                             ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                            ▊▊    ▊▊▊▊▊▊    ▊     ▊▊                            ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                           ▊▊     ▊         ▊      ▊▊                           ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                          ▊▊      ▊         ▊       ▊▊                          ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                         ▊▊       ▊         ▊        ▊▊                         ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                         ▊▊▊▊▊▊▊▊ ▊▊▊▊▊▊▊▊ ▊▊▊        ▊                         ")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
			rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
		}

		rt.Out.Put(out.Opts{Type: out.Info, Typewrite: 25},
			"%s %s",
			rt.Out.Stylize(
				out.Style{FG: out.ColorPrimary, BG: out.ColorSecondary},
				"zeit"),
			rt.Build.Version,
		)
		rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 25},
			" %s %s",
			rt.Out.FG(out.ColorSecondary, "Commit:"),
			rt.Build.Commit,
		)
		rt.Out.Put(out.Opts{Type: out.Plain, Typewrite: 25},
			" %s %s",
			rt.Out.FG(out.ColorSecondary, "Build date:"),
			rt.Build.Date,
		)
	},
}

func init() {
}
