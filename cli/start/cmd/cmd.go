package startCmd

import (
	"encoding/json"

	"github.com/mrusme/zeit/cli/start/shared"
	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/models/project"
	"github.com/mrusme/zeit/models/task"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

const (
	FormatUnspecified = ""
	FormatCLI         = "cli"
	FormatJSON        = "json"
)

var (
	flagFormat string
	flags      *argsparser.ParsedArgs
)

var aliasMap = runtime.AliasMap{
	"start":  {"started", "sta", "str", "s"},
	"switch": {"switched", "switch", "sw"},
	"resume": {"resume", "re"},
}

var Cmd = &cobra.Command{
	Use:               "start [flags] [arguments]",
	Aliases:           aliasMap.GetAliases(),
	Short:             "zeit start",
	Long:              "Start tracking",
	Example:           "zeit start work with note \"Hello World\" on myproject/mytask",
	ValidArgsFunction: shared.DynamicArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var pargs *argsparser.ParsedArgs
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		calledAs := rt.GetCommandCall(cmd)
		cmdName := aliasMap.GetCommandNameForAlias(calledAs)

		pargs, err = argsparser.POP("end", flags, args, rt.Logger)
		rt.NilOrDie(err)

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
					b.ProjectSID,
					b.TaskSID,
				)
				rt.NilOrDie(err)
			}
		}

		var nb *block.Block
		switch cmdName {
		case "start":
			nb, err = block.Start(rt.Database, b)
			rt.NilOrDie(err)
		case "switch":
			nb, err = block.Switch(rt.Database, b)
			rt.NilOrDie(err)
		case "resume":
			nb, err = block.Resume(rt.Database, b)
			rt.NilOrDie(err)
		}

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, pargs, cmdName, nb)
		case FormatCLI:
			outputCLI(rt, pargs, cmdName, nb)
		case FormatJSON:
			outputJSON(rt, pargs, cmdName, nb)
		}
		return
	},
}

func outputCLI(
	rt *runtime.Runtime,
	pargs *argsparser.ParsedArgs,
	cmdName string,
	nb *block.Block,
) {
	switch cmdName {
	case "start":
		if nb.TimestampEnd.IsZero() == true {
			rt.Out.Put(out.Opts{Type: out.Start},
				"Started tracking on %s ...",
				rt.Out.Stylize(out.Style{FG: out.ColorPrimary},
					"%s/%s", nb.ProjectSID, nb.TaskSID),
			)
		} else {
			rt.Out.Put(out.Opts{Type: out.Start},
				"Tracked on %s.",
				rt.Out.Stylize(out.Style{FG: out.ColorPrimary},
					"%s/%s", nb.ProjectSID, nb.TaskSID),
			)
		}
	case "switch":
		rt.Out.Put(out.Opts{Type: out.Start},
			"Switched tracking to %s ...",
			rt.Out.Stylize(out.Style{FG: out.ColorPrimary},
				"%s/%s", nb.ProjectSID, nb.TaskSID),
		)
	case "resume":
		rt.Out.Put(out.Opts{Type: out.Resume},
			"Resumed tracking on %s ...",
			rt.Out.Stylize(out.Style{FG: out.ColorPrimary},
				"%s/%s", nb.ProjectSID, nb.TaskSID),
		)
	}
}

func outputJSON(
	rt *runtime.Runtime,
	pargs *argsparser.ParsedArgs,
	cmdName string,
	nb *block.Block,
) {
	var statusOut *out.StatusOut

	statusOut = new(out.StatusOut)
	statusOut.IsRunning = true
	statusOut.ProjectSID = nb.ProjectSID
	statusOut.TaskSID = nb.TaskSID

	switch cmdName {
	case "start":
		statusOut.Status = "started"
	case "switch":
		statusOut.Status = "switched"
	case "resume":
		statusOut.Status = "resumed"
	}

	prettyJSON, err := json.MarshalIndent(statusOut, "", "  ")
	rt.NilOrDie(err)

	rt.Out.Put(out.Opts{Type: out.Plain}, "%s", string(prettyJSON))
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

	Cmd.PersistentFlags().StringVarP(
		&flagFormat,
		"format",
		"f",
		"",
		"Output format (cli, json) (default \"cli\")",
	)
}
