package cmd

import (
	"encoding/json"
	"os"
	"time"

	blockCmd "xn--gckvb8fzb.com/zeit/cli/block/cmd"
	endCmd "xn--gckvb8fzb.com/zeit/cli/end/cmd"
	exportCmd "xn--gckvb8fzb.com/zeit/cli/export/cmd"
	importCmd "xn--gckvb8fzb.com/zeit/cli/import/cmd"
	projectCmd "xn--gckvb8fzb.com/zeit/cli/project/cmd"
	startCmd "xn--gckvb8fzb.com/zeit/cli/start/cmd"
	statCmd "xn--gckvb8fzb.com/zeit/cli/stat/cmd"
	taskCmd "xn--gckvb8fzb.com/zeit/cli/task/cmd"
	versionCmd "xn--gckvb8fzb.com/zeit/cli/version/cmd"
	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/runtime"
	"github.com/spf13/cobra"
)

const (
	FormatUnspecified = ""
	FormatCLI         = "cli"
	FormatJSON        = "json"
)

var (
	flagDebug  bool
	flagColor  string
	flagFormat string
)

var rootCmd = &cobra.Command{
	Use:   "zeit",
	Short: "A command line time tracker.",
	Long: "Zeit, erfassen. A command line tool for tracking time spent on " +
		" activities.\n\n",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), true)
		defer rt.End()

		found, _, b, err := block.GetActive(rt.Database)
		rt.NilOrDie(err)

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, found, b)
		case FormatCLI:
			outputCLI(rt, found, b)
		case FormatJSON:
			outputJSON(rt, found, b)
		}
	},
}

func outputCLI(
	rt *runtime.Runtime,
	found bool,
	b *block.Block,
) {
	if found == true {
		duration := time.Now().Sub(b.TimestampStart)
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		seconds := int(duration.Seconds()) % 60

		rt.Out.Put(out.Opts{Type: out.Start},
			"Tracking on %s/%s for %s",
			rt.Out.Stylize(
				out.Style{FG: out.ColorPrimary},
				"%s", b.ProjectSID),
			rt.Out.Stylize(
				out.Style{FG: out.ColorPrimary},
				"%s", b.TaskSID),
			rt.Out.Stylize(
				out.Style{FG: out.ColorCyan},
				"%02d:%02d:%02d", hours, minutes, seconds),
		)
	} else {
		rt.Out.Put(out.Opts{Type: out.End}, "Not tracking")
	}
}

func outputJSON(
	rt *runtime.Runtime,
	found bool,
	b *block.Block,
) {
	var statusOut *out.StatusOut

	statusOut = new(out.StatusOut)

	if found == true {
		statusOut.IsRunning = true
		statusOut.ProjectSID = b.ProjectSID
		statusOut.TaskSID = b.TaskSID
		statusOut.Timer = int64(time.Now().Sub(b.TimestampStart).Seconds())
		statusOut.Status = "tracking"
	} else {
		statusOut.IsRunning = false
		statusOut.Status = "not tracking"
	}

	prettyJSON, err := json.MarshalIndent(statusOut, "", "  ")
	rt.NilOrDie(err)

	rt.Out.Put(out.Opts{Type: out.Plain}, "%s", string(prettyJSON))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = runtime.NewBuild().Version

	rootCmd.AddCommand(startCmd.Cmd)
	rootCmd.AddCommand(endCmd.Cmd)
	rootCmd.AddCommand(projectCmd.Cmd)
	rootCmd.AddCommand(taskCmd.Cmd)
	rootCmd.AddCommand(blockCmd.Cmd)
	rootCmd.AddCommand(importCmd.Cmd)
	rootCmd.AddCommand(exportCmd.Cmd)
	rootCmd.AddCommand(statCmd.Cmd)
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

	rootCmd.PersistentFlags().StringVarP(
		&flagFormat,
		"format",
		"f",
		"",
		"Output format (cli, json) (default \"cli\")",
	)
}
