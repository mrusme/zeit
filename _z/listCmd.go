package z

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

var (
	listTotalTime            bool
	listOnlyProjectsAndTasks bool
	listOnlyTasks            bool
	appendProjectIDToTask    bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List activities",
	Long:  "List all tracked activities.",
	Run: func(cmd *cobra.Command, args []string) {
		filteredEntries := listEntries()

		totalHours := decimal.NewFromInt(0)
		for _, entry := range filteredEntries {
			totalHours = totalHours.Add(entry.GetDuration())
			fmt.Printf("%s\n", entry.GetOutput(false))
		}

		if listTotalTime == true {
			fmt.Printf("\nTOTAL: %s H\n\n", fmtHours(totalHours))
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&since, "since", "", "Date/time to start the list from")
	listCmd.Flags().StringVar(&until, "until", "", "Date/time to list until")
	listCmd.Flags().StringVar(&listRange, "range", "", "Shortcut for --since and --until that accepts: "+strings.Join(Ranges(), ", "))
	listCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be listed")
	listCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be listed")
	listCmd.Flags().BoolVar(&fractional, "decimal", false, "Show fractional hours in decimal format instead of minutes")
	listCmd.Flags().BoolVar(&listTotalTime, "total", false, "Show total time of hours for listed activities")
	listCmd.Flags().BoolVar(&listOnlyProjectsAndTasks, "only-projects-and-tasks", false, "Only list projects and their tasks, no entries")
	listCmd.Flags().BoolVar(&listOnlyTasks, "only-tasks", false, "Only list tasks, no projects nor entries")
	listCmd.Flags().BoolVar(&appendProjectIDToTask, "append-project-id-to-task", false, "Append project ID to tasks in the list")

	flagName := "task"
	listCmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		user := GetCurrentUser()
		entries, _ := database.ListEntries(user)
		_, tasks := listProjectsAndTasks(entries)
		return tasks, cobra.ShellCompDirectiveDefault
	})
}
