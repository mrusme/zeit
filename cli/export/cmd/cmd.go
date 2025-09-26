package exportCmd

import (
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:       "export",
	Aliases:   []string{"ex", "x", "dump"},
	Short:     "zeit export",
	Long:      "Export the zeit database to various formats",
	Example:   "zeit export myproject/mytask from 2 days ago until now ",
	ValidArgs: []string{"from", "until"},
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		all, err := rt.Database.GetAllRowsAsBytes()
		if err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		for key, row := range all {
			rt.Out.Put(out.Opts{Type: out.Info},
				"%s %s",
				rt.Out.Stylize(out.Style{FG: out.ColorPrimary},
					"%s", key),
				row,
			)
		}
	},
}

func init() {
}
