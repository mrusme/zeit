package projectEditCmd

import (
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/project"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var (
	flagDisplayName string
	flagColor       string
)

var Cmd = &cobra.Command{
	Use:     "edit [flags] project-sid",
	Aliases: []string{},
	Short:   "zeit project edit",
	Long:    "Edit zeit projects",
	Example: "zeit project edit myproject",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var pj *project.Project
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		pj, err = project.GetBySID(rt.Database, args[0])
		rt.NilOrDie(err)

		if flagDisplayName != "" {
			pj.DisplayName = flagDisplayName
		}
		if flagColor != "" {
			pj.Color = flagColor
		}

		err = project.Set(rt.Database, pj)
		rt.NilOrDie(err)

		rt.Out.Put(out.Opts{Type: out.Ok}, "Project updated!")
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
