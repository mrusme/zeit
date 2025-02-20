package z

import (
	"github.com/spf13/cobra"
)

var switchBackCmd = &cobra.Command{
	Use:   "switchback",
	Short: "switchback to the task before the last one",
	Long:  "End running activity and resume the task which was before, which can either be kept running until 'finish' is being called or parameterized to be a finished activity.",
	Run: func(cmd *cobra.Command, args []string) {
		finish = switchString
		finishTask(FinishOnlyTime)

		finish = ""
		begin = switchString
		resumeTask(2)
	},
}

func init() {
	rootCmd.AddCommand(switchBackCmd)

	switchBackCmd.Flags().StringVarP(&switchString, "begin", "b", "", "Time the new activity should begin at and the old one ends\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
}
