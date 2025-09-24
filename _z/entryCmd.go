package z

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var entryCmd = &cobra.Command{
	Use:   "entry ([flags]) [id]",
	Short: "Display or update activity",
	Long:  "Display or update tracked activity.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		user := GetCurrentUser()
		id := args[0]

		entry, err := database.GetEntry(user, id)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}

		if begin != "" || finish != "" || project != "" || notes != "" || task != "" {
			if begin != "" {
				entry.Begin, err = entry.SetBeginFromString(begin, entry.Begin)
				if err != nil {
					fmt.Printf("%s %+v\n", CharError, err)
					os.Exit(1)
				}
			}

			if finish != "" {
				entry.Finish, err = entry.SetFinishFromString(finish, entry.Finish)
				if err != nil {
					fmt.Printf("%s %+v\n", CharError, err)
					os.Exit(1)
				}
			}

			if project != "" {
				entry.Project = project
			}

			if task != "" {
				entry.Task = task
			}

			if notes != "" {
				entry.Notes = strings.Replace(notes, "\\n", "\n", -1)
			}

			_, err = database.UpdateEntry(user, entry)
			if err != nil {
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		}

		fmt.Printf("%s %s\n", CharInfo, entry.GetOutput(true))
		return
	},
}

func init() {
	rootCmd.AddCommand(entryCmd)
	entryCmd.Flags().StringVarP(&begin, "begin", "b", "", "Update date/time the activity began at")
	entryCmd.Flags().StringVarP(&finish, "finish", "s", "", "Update date/time the activity finished at")
	entryCmd.Flags().StringVarP(&project, "project", "p", "", "Update activity project")
	entryCmd.Flags().StringVarP(&notes, "notes", "n", "", "Update activity notes")
	entryCmd.Flags().StringVarP(&task, "task", "t", "", "Update activity task")
	entryCmd.Flags().BoolVar(&fractional, "decimal", false, "Show fractional hours in decimal format instead of minutes")

	flagName := "task"
	entryCmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		user := GetCurrentUser()
		entries, _ := database.ListEntries(user)
		_, tasks := listProjectsAndTasks(entries)
		return tasks, cobra.ShellCompDirectiveDefault
	})
}
