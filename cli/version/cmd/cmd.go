package versionCmd

import (
	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/runtime"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "zeit version",
	Long:  "Display zeit version information",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		outputVersion(out.New(runtime.GetOutputColor(cmd)), runtime.NewBuild())
	},
}

func outputVersion(o *out.Out, build runtime.Build) {
	if o.InColor() {
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ █████████████               ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ █████████████               ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ █████████████               ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ █████████████               ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                     ████████ ███████        ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                    ███████   ████████████   ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                   ███████    ████████████   ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                  ███████     ████████████   ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                 ███████      ████████████   ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "                ███████       ███████        ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "               ██████████████ ██████████████ ██████ ███████                     ")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
		o.Put(out.Opts{Type: out.Plain, Typewrite: 10}, "")
	}

	o.Put(out.Opts{Type: out.Info, Typewrite: 25},
		"%s %s",
		o.Stylize(
			out.Style{FG: out.ColorPrimary, BG: out.ColorSecondary},
			"zeit"),
		build.Version,
	)
	o.Put(out.Opts{Type: out.Plain, Typewrite: 25},
		"  %s %s",
		o.FG(out.ColorSecondary, "Commit:"),
		build.Commit,
	)
	o.Put(out.Opts{Type: out.Plain, Typewrite: 25},
		"  %s %s",
		o.FG(out.ColorSecondary, "Build date:"),
		build.Date,
	)
}

func init() {
}
