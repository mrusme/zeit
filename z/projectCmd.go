package z

import (
  "os"
  "fmt"
  // "time"
  "github.com/spf13/cobra"
  // "github.com/gookit/color"
)

var projectColor string

var projectCmd = &cobra.Command{
  Use:   "project ([flags]) [project]",
  Short: "Project settings",
  Long: "Configure project settings.",
  Args: cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    user := GetCurrentUser()
    projectName := args[0]

    project, err := database.GetProject(user, projectName)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    project.Name = projectName

    if projectColor != "" {
      project.Color = projectColor
    }

    err = database.UpdateProject(user, projectName, project)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    fmt.Printf("%s project updated\n", CharInfo)
    return
  },
}

func init() {
  rootCmd.AddCommand(projectCmd)
  projectCmd.Flags().StringVarP(&projectColor, "color", "c", "", "Set the color of the project (hex code, e.g. #121212)")

  var err error
  database, err = InitDatabase()
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }
}
