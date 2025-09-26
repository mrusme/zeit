package endCmd

import (
	"fmt"

	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var flags *argsparser.ParsedArgs

var Cmd = &cobra.Command{
	Use:       "end [flags] [arguments]",
	Aliases:   []string{"ended", "en", "e"},
	Short:     "zeit end",
	Long:      "End tracking",
	Example:   "zeit end with note \"Issue ID 123\" 5 minutes ago",
	ValidArgs: []string{"block", "work", "with"},
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		pargs, err := argsparser.Parse("end", args)
		if err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		pargs.OverrideWith(flags)

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

		if err = block.End(rt, b); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		rt.Out.Put(out.Opts{Type: out.End}, "Ended tracking")
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
