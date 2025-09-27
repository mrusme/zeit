package exportCmd

import (
	"encoding/json"
	"strings"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/argsparser"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/activeblock"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/models/config"
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
	flags      *argsparser.ParsedArgs
	flagFormat string
	flagBackup bool
)

var Cmd = &cobra.Command{
	Use:       "export",
	Aliases:   []string{"ex", "x", "dump"},
	Short:     "zeit export",
	Long:      "Export the zeit database to various formats",
	Example:   "zeit export all of myproject/mytask from 2 days ago until now",
	ValidArgs: []string{"from", "until"},
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		var pargs *argsparser.ParsedArgs
		var blockMap map[string]*block.Block = make(map[string]*block.Block)
		var dump map[string]interface{} = make(map[string]interface{})

		var err error
		var keys []string

		flagFormat = strings.ToLower(flagFormat)

		if flagBackup == false {
			pargs, err = argsparser.Parse("export", args)
			rt.NilOrDie(err)

			pargs.OverrideWith(flags)

			rt.Logger.Debug("Parsed args",
				"pargs", pargs,
				"GetTimestampStart", pargs.GetTimestampStart(),
				"GetTimestampEnd", pargs.GetTimestampEnd(),
			)

			err := pargs.Process()
			rt.NilOrDie(err)

			rt.Logger.Debug("Processed args",
				"pargs", pargs,
				"GetTimestampStart", pargs.GetTimestampStart(),
				"GetTimestampEnd", pargs.GetTimestampEnd(),
			)
		}

		blockMap, err = block.List(rt.Database)
		rt.NilOrDie(err)

		var filterByTimestamp bool = false
		if flagBackup == false &&
			pargs.GetTimestampStart().IsZero() == false &&
			pargs.GetTimestampEnd().IsZero() == false &&
			pargs.GetTimestampStart().Before(pargs.GetTimestampEnd()) {
			filterByTimestamp = true
		}

		for key, b := range blockMap {
			if flagBackup == false && ((filterByTimestamp == true &&
				((b.TimestampStart.Before(pargs.GetTimestampStart()) ||
					b.TimestampStart.After(pargs.GetTimestampEnd())) ||
					(b.TimestampEnd.Before(pargs.GetTimestampStart()) ||
						b.TimestampEnd.After(pargs.GetTimestampEnd())))) ||
				(pargs.ProjectSID != "" && b.ProjectSID != pargs.ProjectSID) ||
				(pargs.TaskSID != "" && b.TaskSID != pargs.TaskSID)) {
				continue
			} else {
				keys = append(keys, key)
				dump[key] = blockMap[key]
			}
		}

		if flagBackup == true {
			if flagFormat == FormatUnspecified {
				flagFormat = FormatJSON
			}

			err = InjectBackupData(rt, &keys, &dump)
			rt.NilOrDie(err)
		}

		database.SortKeys(keys)

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, dump, keys)
		case FormatCLI:
			outputCLI(rt, dump, keys)
		case FormatJSON:
			outputJSON(rt, dump, keys)
		}
	},
}

func InjectBackupData(
	rt *runtime.Runtime,
	keys *[]string,
	dump *map[string]interface{},
) error {
	var cfg *config.Config
	var ab *activeblock.ActiveBlock
	var projectMap map[string]*project.Project = make(map[string]*project.Project)
	var taskMap map[string]*task.Task = make(map[string]*task.Task)
	var err error

	// ------------------------- Static Key Entries ----------------------- //
	if cfg, err = config.Get(rt.Database); err != nil {
		return err
	}

	if ab, err = activeblock.Get(rt.Database); err != nil {
		return err
	}

	*keys = append(*keys,
		config.KEY,
		activeblock.KEY,
	)
	(*dump)[config.KEY] = cfg
	(*dump)[activeblock.KEY] = ab

	// ------------------------------ Projects ---------------------------- //
	if projectMap, err = project.List(rt.Database); err != nil {
		return err
	}

	for key := range projectMap {
		*keys = append(*keys, key)
		(*dump)[key] = projectMap[key]
	}

	// -------------------------------- Tasks ----------------------------- //
	if taskMap, err = task.List(rt.Database); err != nil {
		return err
	}

	for key := range taskMap {
		*keys = append(*keys, key)
		(*dump)[key] = taskMap[key]
	}

	return nil
}

func outputCLI(
	rt *runtime.Runtime,
	dump map[string]interface{},
	sorting []string,
) {
	for _, key := range sorting {
		rt.Out.Put(out.Opts{Type: out.Info},
			"%s %s",
			rt.Out.Stylize(out.Style{FG: out.ColorPrimary},
				"%s", key),
			dump[key],
		)
	}
}

func outputJSON(
	rt *runtime.Runtime,
	dump map[string]interface{},
	sorting []string,
) {
	prettyJSON, err := json.MarshalIndent(dump, "", "  ")
	rt.NilOrDie(err)

	rt.Out.Put(out.Opts{Type: out.Plain}, string(prettyJSON))
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

	Cmd.PersistentFlags().StringVarP(
		&flagFormat,
		"format",
		"f",
		"",
		"Export format (cli, json) (default \"cli\")",
	)
	Cmd.PersistentFlags().BoolVarP(
		&flagBackup,
		"backup",
		"b",
		false,
		"Export entire database as backup",
	)
}
