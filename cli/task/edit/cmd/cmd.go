package taskEditCmd

import (
	"strings"

	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/task"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var (
	flagDisplayName string
	flagColor       string
)

var Cmd = &cobra.Command{
	Use:     "edit [flags] project-sid[/]task-sid",
	Aliases: []string{},
	Short:   "zeit task edit",
	Long:    "Edit zeit tasks",
	Example: "zeit task edit myproject mytask",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var pj *task.Task
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		var projectSID string
		var taskSID string
		var found bool

		if len(args) == 1 {
			projectSID, taskSID, found = strings.Cut(args[0], "/")
			if found == false {
				rt.Out.Put(out.Opts{Type: out.Error}, "Please provide a project and "+
					"task SID in the format myproject/mytask or myproject mytask")
			}
		} else if len(args) == 2 {
			projectSID = args[0]
			taskSID = args[1]
		}

		pj, err = task.GetBySID(rt.Database, projectSID, taskSID)
		rt.NilOrDie(err)

		if flagDisplayName != "" {
			pj.DisplayName = flagDisplayName
		}
		if flagColor != "" {
			pj.Color = flagColor
		}

		err = task.Set(rt.Database, pj)
		rt.NilOrDie(err)

		rt.Out.Put(out.Opts{Type: out.Ok}, "Task updated!")
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(
		&flagDisplayName,
		"display-name",
		"d",
		"",
		"Set the display name",
	)
	Cmd.PersistentFlags().StringVarP(
		&flagColor,
		"color",
		"c",
		"",
		"Set the color",
	)
}
