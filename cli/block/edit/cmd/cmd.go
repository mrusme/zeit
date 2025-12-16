package blockEditCmd

import (
	"strings"

	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var flags *argsparser.ParsedArgs = &argsparser.ParsedArgs{}

var Cmd = &cobra.Command{
	Use:     "edit [flags] key",
	Aliases: []string{},
	Short:   "zeit block edit",
	Long:    "Edit zeit blocks",
	Example: "zeit block edit 01998b32-7f89-7373-a192-56417e0bc89f",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var pargs *argsparser.ParsedArgs
		var b *block.Block
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), false)
		defer rt.End()

		var blockKey string = args[0]
		if strings.Index(blockKey, "block:") == -1 {
			blockKey = "block:" + blockKey
		}

		pargs, err = argsparser.POP("edit", flags, []string{}, rt.Logger)
		rt.NilOrDie(err)

		b, err = block.Get(rt.Database, blockKey)
		rt.NilOrDie(err)

		err = b.FromProcessedArgs(pargs)

		err = block.Set(rt.Database, b)
		rt.NilOrDie(err)

		rt.Out.Put(out.Opts{Type: out.Ok}, "Block updated!")
	},
}

func init() {
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
