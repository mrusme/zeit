package cmd

import (
	"os"

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
		rt := runtime.New(runtime.GetLogLevel(cmd))
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
