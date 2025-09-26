package startCmd

import (
	"fmt"

	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
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
		if err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		pargs.OverrideWith(flags)

		fmt.Printf("Project ID: %s\n",
			pargs.ProjectSID)
		fmt.Printf("Task ID: %s\n",
			pargs.TaskSID)
		fmt.Printf("Note: %s\n",
			pargs.Note)
		fmt.Printf("Start Timestamp: %s\n",
			pargs.TimestampStart)
		fmt.Printf("End Timestamp: %s\n",
			pargs.TimestampEnd)

		if err := pargs.Process(); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		fmt.Printf("Start Timestamp (time): %s\n",
			pargs.GetTimestampStart())
		fmt.Printf("End Timestamp (time): %s\n",
			pargs.GetTimestampEnd())

		b, err := block.New(rt.Config.UserKey)
		if err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		if err = b.FromProcessedArgs(pargs); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		switch cmdName {
		case "start":
			if err = block.Start(rt, b); err != nil {
				rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
				rt.Exit(1)
			}

			rt.Out.Put(out.Opts{Type: out.Start}, "Started tracking ...")
		case "switch":
			if err = block.Switch(rt, b); err != nil {
				rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
				rt.Exit(1)
			}

			rt.Out.Put(out.Opts{Type: out.Start}, "Switched tracking ...")
		case "resume":
			if err = block.Resume(rt, b); err != nil {
				rt.Out.Put(out.Opts{Type: out.Switch}, err.Error())
				rt.Exit(1)
			}

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
