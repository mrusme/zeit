package endCmd

import (
	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var flags *argsparser.ParsedArgs = &argsparser.ParsedArgs{}

var aliasMap = runtime.AliasMap{
	"end":   {"en", "e"},
	"stop":  {"stop", "sto", "stp"},
	"pause": {"pause", "ps", "p"},
}

var Cmd = &cobra.Command{
	Use:       "end [flags] [arguments]",
	Aliases:   aliasMap.GetAliases(),
	Short:     "zeit end",
	Long:      "End tracking",
	Example:   "zeit end with note \"Issue ID 123\" 5 minutes ago",
	ValidArgs: []string{"block", "work", "with"},
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

		err = block.End(rt.Database, b)
		rt.NilOrDie(err)

		switch cmdName {
		case "end":
			rt.Out.Put(out.Opts{Type: out.End}, "Ended tracking")
		case "stop":
			rt.Out.Put(out.Opts{Type: out.End}, "Stopped tracking")
		case "pause":
			rt.Out.Put(out.Opts{Type: out.Pause}, "Paused tracking")
		}
		return
	},
}

func init() {
	flags = new(argsparser.ParsedArgs)

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
