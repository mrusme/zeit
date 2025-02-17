package z

import (
  "github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
  Use:   "switch",
  Short: "switch to another task",
  Long:  "End running activity and track new activity, which can either be kept running until 'finish' is being called or parameterized to be a finished activity.",
  Run: func(cmd *cobra.Command, args []string) {

    finish = switchString
    finishTask(FinishOnlyTime)

    finish = ""
    begin = switchString
    trackTask()
  },
}

func init() {
  rootCmd.AddCommand(switchCmd)

  switchCmd.Flags().StringVarP(&switchString, "begin", "b", "", "Time the new activity should begin at and the old one ends\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).")
  switchCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be assigned")
  switchCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be assigned")
  switchCmd.Flags().StringVarP(&notes, "notes", "n", "", "Activity notes")

  flagName := "task"
  switchCmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    user := GetCurrentUser()
    entries, _ := database.ListEntries(user)
    _, tasks := listProjectsAndTasks(entries)
    return tasks, cobra.ShellCompDirectiveDefault
  })
}

