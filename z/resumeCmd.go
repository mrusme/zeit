package z

import (
  "github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
  Use:   "resume",
  Short: "Resume last task",
  Long:  "Track new activity with all parameters of the last task (based on begin time)",
  Run: func(cmd *cobra.Command, args []string) {
    resumeTask(1)
  },
}

func init() {
  rootCmd.AddCommand(resumeCmd)

  resumeCmd.Flags().StringVarP(&begin, "begin", "b", "", "Time the activity should begin at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
  resumeCmd.Flags().StringVarP(&finish, "finish", "s", "", "Time the activity should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
}
