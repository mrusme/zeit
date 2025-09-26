package exportCmd

import (
	"encoding/json"
	"fmt"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var (
	flags   *argsparser.ParsedArgs
	flagAll bool
)

var Cmd = &cobra.Command{
	Use:       "export",
	Aliases:   []string{"ex", "x", "dump"},
	Short:     "zeit export",
	Long:      "Export the zeit database to various formats",
	Example:   "zeit export myproject/mytask from 2 days ago until now",
	ValidArgs: []string{"from", "until"},
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		var byteMap map[string][]byte
		var blockMap map[string]*block.Block = make(map[string]*block.Block)
		var err error
		var keys []string

		if flagAll == true {
			byteMap, err = rt.Database.GetAllRowsAsBytes()
			if err != nil {
				rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
				rt.Exit(1)
			}

			for key := range byteMap {
				keys = append(keys, key)
			}
		} else {
			pargs, err := argsparser.Parse("export", args)
			if err != nil {
				rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
				rt.Exit(1)
			}

			pargs.OverrideWith(flags)

			fmt.Printf("Project ID: %s\n",
				pargs.ProjectSID)
			fmt.Printf("Task ID: %s\n",
				pargs.TaskSID)
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

			err = database.GetPrefixedRowsAsStruct(rt.Database, "block:", blockMap)
			if err != nil {
				rt.Out.Put(out.Opts{Type: out.Error}, err.Error())
				rt.Exit(1)
			}

			var filterByTimestamp bool = false
			if pargs.GetTimestampStart().IsZero() == false &&
				pargs.GetTimestampEnd().IsZero() == false &&
				pargs.GetTimestampStart().Before(pargs.GetTimestampEnd()) {
				filterByTimestamp = true
			}

			for key, b := range blockMap {
				if filterByTimestamp == true &&
					((b.TimestampStart.Before(pargs.GetTimestampStart()) ||
						b.TimestampStart.After(pargs.GetTimestampEnd())) ||
						(b.TimestampEnd.Before(pargs.GetTimestampStart()) ||
							b.TimestampEnd.After(pargs.GetTimestampEnd()))) {
					continue
				} else {
					keys = append(keys, key)
				}
			}

		}

		database.SortKeys(keys)

		for _, key := range keys {
			var content string

			if flagAll == true {
				content = string(byteMap[key])
			} else {
				tmp, err := json.Marshal(blockMap[key])
				if err != nil {
					tmp = []byte("{\"error\":\"Marshal failed\"}")
				}
				content = string(tmp)
			}

			rt.Out.Put(out.Opts{Type: out.Info},
				"%s %s",
				rt.Out.Stylize(out.Style{FG: out.ColorPrimary},
					"%s", key),
				content,
			)
		}
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

	Cmd.PersistentFlags().BoolVarP(
		&flagAll,
		"all",
		"a",
		false,
		"Export entire database",
	)
}
