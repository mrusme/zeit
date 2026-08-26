package exportCmd

import (
	"encoding/json"
	"strings"
	"time"

	"xn--gckvb8fzb.com/zeit/database"
	"xn--gckvb8fzb.com/zeit/helpers/argsparser"
	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/models/activeblock"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/models/config"
	"xn--gckvb8fzb.com/zeit/models/project"
	"xn--gckvb8fzb.com/zeit/models/task"
	"xn--gckvb8fzb.com/zeit/runtime"
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
	Use:       "export [flags] [arguments]",
	Aliases:   []string{"ex", "x", "dump"},
	Short:     "zeit export",
	Long:      "Export the zeit database to various formats",
	Example:   "zeit export all of myproject/mytask from 2 days ago until now",
	ValidArgs: []string{"from", "until"},
	Run: func(cmd *cobra.Command, args []string) {
		var pargs *argsparser.ParsedArgs
		var blockMap map[string]*block.Block = make(map[string]*block.Block)
		var dump map[string]interface{} = make(map[string]interface{})
		var keys []string
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), true)
		defer rt.End()

		flagFormat = strings.ToLower(flagFormat)

		if flagBackup == false {
			pargs, err = argsparser.POP("export", flags, args, rt.Logger)
			rt.NilOrDie(err)
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
	for idx, key := range sorting {
		switch model := dump[key].(type) {
		case *block.Block:
			outputBlockCLI(rt, key, model)
		case *project.Project:
			outputProjectCLI(rt, key, model)
		case *task.Task:
			outputTaskCLI(rt, key, model)
		case *config.Config:
			outputConfigCLI(rt, key, model)
		case *activeblock.ActiveBlock:
			outputActiveBlockCLI(rt, key, model)
		}

		if idx < len(sorting)-1 {
			rt.Out.Put(out.Opts{Type: out.Plain}, "")
		}
	}
}

func outputBlockCLI(rt *runtime.Runtime, key string, b *block.Block) {
	var duration time.Duration
	if b.TimestampStart.IsZero() == false &&
		b.TimestampEnd.IsZero() == false {
		duration = b.TimestampEnd.Sub(b.TimestampStart)
	}

	rt.Out.Put(out.Opts{Type: out.Info},
		"%s\n  %s %s %s  %s %s\n  tracked on %s\n  %s",
		rt.Out.Stylize(
			out.Style{FG: out.ColorPrimary},
			"%s", key,
		),
		rt.Out.Stylize(
			out.Style{FG: out.OutputPrefixes[out.Start].Color},
			"%s%s",
			out.OutputPrefixes[out.Start].Char,
			b.TimestampStart.Format(time.DateTime),
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorSecondary},
			"→",
		),
		rt.Out.Stylize(
			out.Style{FG: out.OutputPrefixes[out.End].Color},
			"%s%s",
			out.OutputPrefixes[out.End].Char,
			out.EndTimestamp(b.TimestampEnd),
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorSecondary},
			"⭘",
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorWhite},
			"%s", out.Seconds(duration),
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorPrimary},
			"%s/%s", b.ProjectSID, b.TaskSID,
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorBrightBlack},
			"%s", block.GetNotePreview(b.Note, 0),
		),
	)
}

func outputProjectCLI(rt *runtime.Runtime, key string, pj *project.Project) {
	rt.Out.Put(out.Opts{Type: out.Info},
		"%s\n  %s%s",
		rt.Out.Stylize(
			out.Style{FG: out.ColorPrimary},
			"%s", key,
		),
		rt.Out.Stylize(
			out.Style{BG: out.Color(pj.Color), FG: out.ColorBrightWhite, PX: 1},
			"%s", pj.DisplayName,
		),
		rt.Out.Stylize(
			out.Style{BG: out.Color(pj.Color), FG: out.ColorBlack, PX: 1},
			"[%s]", pj.SID,
		),
	)
}

func outputTaskCLI(rt *runtime.Runtime, key string, tk *task.Task) {
	rt.Out.Put(out.Opts{Type: out.Info},
		"%s\n  %s%s",
		rt.Out.Stylize(
			out.Style{FG: out.ColorPrimary},
			"%s", key,
		),
		rt.Out.Stylize(
			out.Style{BG: out.Color(tk.Color), FG: out.ColorBrightWhite, PX: 1},
			"%s", tk.DisplayName,
		),
		rt.Out.Stylize(
			out.Style{BG: out.Color(tk.Color), FG: out.ColorBlack, PX: 1},
			"[%s/%s]", tk.ProjectSID, tk.SID,
		),
	)
}

func outputConfigCLI(rt *runtime.Runtime, key string, cfg *config.Config) {
	rt.Out.Put(out.Opts{Type: out.Info},
		"%s\n  %s %s",
		rt.Out.Stylize(
			out.Style{FG: out.ColorPrimary},
			"%s", key,
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorSecondary},
			"user key",
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorWhite},
			"%s", cfg.UserKey,
		),
	)
}

func outputActiveBlockCLI(
	rt *runtime.Runtime,
	key string,
	ab *activeblock.ActiveBlock,
) {
	rt.Out.Put(out.Opts{Type: out.Info},
		"%s\n  %s %s\n  %s %s",
		rt.Out.Stylize(
			out.Style{FG: out.ColorPrimary},
			"%s", key,
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorSecondary},
			"active  ",
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorWhite},
			"%s", ab.GetActiveBlockKey(),
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorSecondary},
			"previous",
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorWhite},
			"%s", ab.GetPreviousBlockKey(),
		),
	)
}

func outputJSON(
	rt *runtime.Runtime,
	dump map[string]interface{},
	sorting []string,
) {
	prettyJSON, err := json.MarshalIndent(dump, "", "  ")
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
	Cmd.PersistentFlags().BoolVarP(
		&flagBackup,
		"backup",
		"b",
		false,
		"Export entire database as backup",
	)
}
