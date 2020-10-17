package z

import (
  "os"
  "fmt"
  // "time"
  "github.com/spf13/cobra"
  // "github.com/gookit/color"
)

var taskGitRepository string

type Task struct {
  Name          string      `json:"name,omitempty"`
  GitRepository string      `json:"gitRepository,omitempty"`
}

var taskCmd = &cobra.Command{
  Use:   "task ([flags]) [task]",
  Short: "Task settings",
  Long: "Configure task settings.",
  Args: cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()
    taskName := args[0]

    task, err := database.GetTask(user, taskName)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    task.Name = taskName

    if taskGitRepository != "-" {
      task.GitRepository = taskGitRepository
    }

    err = database.UpdateTask(user, taskName, task)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    fmt.Printf("%s task updated\n", CharInfo)
    return
  },
}

func init() {
  rootCmd.AddCommand(taskCmd)
  taskCmd.Flags().StringVarP(&taskGitRepository, "git", "g", "-", "Set the task's Git repository to enable commit message importing into activity notes.\nSet to an empty string '' to remove a previously set repository and disable git log imports.")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
