package cmd

import (
	"fmt"
	"os"
	"time"

	blockCmd "github.com/mrusme/zeit/cli/block/cmd"
	endCmd "github.com/mrusme/zeit/cli/end/cmd"
	exportCmd "github.com/mrusme/zeit/cli/export/cmd"
	projectCmd "github.com/mrusme/zeit/cli/project/cmd"
	startCmd "github.com/mrusme/zeit/cli/start/cmd"
	taskCmd "github.com/mrusme/zeit/cli/task/cmd"
	versionCmd "github.com/mrusme/zeit/cli/version/cmd"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var (
	flagDebug bool
	flagColor string
)

var rootCmd = &cobra.Command{
	Use:   "zeit",
	Short: "A command line time tracker.",
	Long: "Zeit, erfassen. A command line tool for tracking time spent on " +
		" activities.\n\n",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		found, _, b, err := block.GetActive(rt.Database)
		rt.NilOrDie(err)

		if found == true {
			duration := time.Now().Sub(b.TimestampStart)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			seconds := int(duration.Seconds()) % 60

			rt.Out.Put(out.Opts{Type: out.Start},
				"Tracking on %s/%s for %s",
				rt.Out.Stylize(
					out.Style{FG: out.ColorPrimary},
					b.ProjectSID),
				rt.Out.Stylize(
					out.Style{FG: out.ColorPrimary},
					b.TaskSID),
				rt.Out.Stylize(
					out.Style{FG: out.ColorCyan},
					fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)),
			)
		} else {
			rt.Out.Put(out.Opts{Type: out.End}, "Not tracking")
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Errorf("%s\n", err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startCmd.Cmd)
	rootCmd.AddCommand(endCmd.Cmd)
	rootCmd.AddCommand(projectCmd.Cmd)
	rootCmd.AddCommand(taskCmd.Cmd)
	rootCmd.AddCommand(blockCmd.Cmd)
	rootCmd.AddCommand(exportCmd.Cmd)
	rootCmd.AddCommand(versionCmd.Cmd)

	rootCmd.PersistentFlags().BoolVar(
		&flagDebug,
		"debug",
		false,
		"Display debugging output in the console",
	)
	rootCmd.PersistentFlags().StringVar(
		&flagColor,
		"color",
		"auto",
		"When to display icons (always, auto, never)",
	)
}
