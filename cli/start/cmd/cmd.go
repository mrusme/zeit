package startCmd

import (
	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/models/project"
	"github.com/mrusme/zeit/models/task"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var flags *argsparser.ParsedArgs

var aliasMap = runtime.AliasMap{
	"start":  {"started", "sta", "str", "s"},
	"switch": {"switched", "switch", "sw"},
	"resume": {"resume", "re"},
}

var Cmd = &cobra.Command{
	Use:       "start [flags] [arguments]",
	Aliases:   aliasMap.GetAliases(),
	Short:     "zeit start",
	Long:      "Start tracking",
	Example:   "zeit start work with note \"Hello World\" on myproject/mytask",
	ValidArgs: []string{"block", "work", "on", "to", "with", "end"},
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		calledAs := rt.GetCommandCall(cmd)
		cmdName := aliasMap.GetCommandNameForAlias(calledAs)

		pargs, err := argsparser.Parse("start", args)
		rt.NilOrDie(err)

		pargs.OverrideWith(flags)

		rt.Logger.Debug("Parsed args",
			"pargs", pargs,
			"GetTimestampStart", pargs.GetTimestampStart(),
			"GetTimestampEnd", pargs.GetTimestampEnd(),
		)

		err = pargs.Process()
		rt.NilOrDie(err)

		rt.Logger.Debug("Processed args",
			"pargs", pargs,
			"GetTimestampStart", pargs.GetTimestampStart(),
			"GetTimestampEnd", pargs.GetTimestampEnd(),
		)

		b, err := block.New(rt.Config.UserKey)
		rt.NilOrDie(err)

		err = b.FromProcessedArgs(pargs)
		rt.NilOrDie(err)

		if cmdName != "resume" {
			if b.ProjectSID != "" {
				// Insert new project if it doesn't exist yet
				_, err = project.InsertIfNone(
					rt.Database,
					rt.Config.UserKey,
					b.ProjectSID,
				)
				rt.NilOrDie(err)
			}

			if b.TaskSID != "" {
				// Insert new task if it doesn't exist yet
				_, err = task.InsertIfNone(
					rt.Database,
					rt.Config.UserKey,
					b.TaskSID,
				)
				rt.NilOrDie(err)
			}
		}

		switch cmdName {
		case "start":
			err = block.Start(rt.Database, b)
			rt.NilOrDie(err)

			rt.Out.Put(out.Opts{Type: out.Start}, "Started tracking ...")
		case "switch":
			err = block.Switch(rt.Database, b)
			rt.NilOrDie(err)

			rt.Out.Put(out.Opts{Type: out.Start}, "Switched tracking ...")
		case "resume":
			err = block.Resume(rt.Database, b)
			rt.NilOrDie(err)

			rt.Out.Put(out.Opts{Type: out.Resume}, "Resumed tracking ...")
		}
		return
	},
}

func init() {
	flags = new(argsparser.ParsedArgs)

	Cmd.PersistentFlags().StringVarP(
		&flags.ProjectSID,
		"project",
		"p",
		"",
		"Project Simplified-ID",
	)
	Cmd.PersistentFlags().StringVarP(
		&flags.TaskSID,
		"task",
		"t",
		"",
		"Task Simplified-ID",
	)
	Cmd.PersistentFlags().StringVarP(
		&flags.Note,
		"note",
		"n",
		"",
		"Note",
	)
	Cmd.PersistentFlags().StringVarP(
		&flags.TimestampStart,
		"start",
		"s",
		"",
		"Start timestamp",
	)
	Cmd.PersistentFlags().StringVarP(
		&flags.TimestampEnd,
		"end",
		"e",
		"",
		"End timestamp",
	)
}
