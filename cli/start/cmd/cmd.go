package startCmd

import (
	"fmt"

	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var (
	flagProjectID      string
	flagTaskID         string
	flagNote           string
	flagTimestampStart string
	flagTimestampEnd   string
)

var Cmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"sta", "str", "s", "resume", "re"},
	Short:   "zeit start",
	Long:    "Start tracking",
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		pargs, err := argsparser.Parse("start", args)
		if err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
			rt.Exit(1)
		}

		fmt.Printf("Project ID: %s\n",
			pargs.ProjectID)
		fmt.Printf("Task ID: %s\n",
			pargs.TaskID)
		fmt.Printf("Note: %s\n",
			pargs.Note)
		fmt.Printf("Start Timestamp: %s\n",
			pargs.TimestampStart)
		fmt.Printf("End Timestamp: %s\n",
			pargs.TimestampEnd)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(
		&flagProjectID,
		"project",
		"p",
		"",
		"Project ID",
	)
	Cmd.PersistentFlags().StringVarP(
		&flagTaskID,
		"task",
		"t",
		"",
		"Task ID",
	)
	Cmd.PersistentFlags().StringVarP(
		&flagNote,
		"note",
		"n",
		"",
		"Note",
	)
	Cmd.PersistentFlags().StringVarP(
		&flagTimestampStart,
		"start",
		"s",
		"",
		"Start timestamp",
	)
	Cmd.PersistentFlags().StringVarP(
		&flagTimestampEnd,
		"end",
		"e",
		"",
		"End timestamp",
	)
}
