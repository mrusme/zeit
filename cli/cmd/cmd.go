package cmd

import (
	"os"

	exportCmd "github.com/mrusme/zeit/cli/export/cmd"
	startCmd "github.com/mrusme/zeit/cli/start/cmd"
	versionCmd "github.com/mrusme/zeit/cli/version/cmd"
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
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startCmd.Cmd)
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
