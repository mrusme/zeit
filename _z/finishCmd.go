package z

import (
	"github.com/spf13/cobra"
)

var finishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Finish currently running activity",
	Long:  "Finishing tracking of currently running activity.",
	Run: func(cmd *cobra.Command, args []string) {
		finishTask(FinishWithMetadata)
	},
}

func init() {
	rootCmd.AddCommand(finishCmd)
	finishCmd.Flags().StringVarP(&begin, "begin", "b", "", "Time the activity should begin at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
	finishCmd.Flags().StringVarP(&finish, "finish", "s", "", "Time the activity should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
	finishCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be assigned")
	finishCmd.Flags().StringVarP(&notes, "notes", "n", "", "Activity notes")
	finishCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be assigned")

	flagName := "task"
	finishCmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		user := GetCurrentUser()
		entries, _ := database.ListEntries(user)
		_, tasks := listProjectsAndTasks(entries)
		return tasks, cobra.ShellCompDirectiveDefault
	})
}
