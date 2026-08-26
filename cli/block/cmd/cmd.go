package blockCmd

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	blockEditCmd "xn--gckvb8fzb.com/zeit/cli/block/edit/cmd"
	"xn--gckvb8fzb.com/zeit/database"
	"xn--gckvb8fzb.com/zeit/helpers/argsparser"
	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/helpers/timestamp"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/runtime"
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

type BlockView struct {
	Key            string           `json:"key"`
	ProjectSID     string           `json:"project_sid"`
	TaskSID        string           `json:"task_sid"`
	Note           string           `json:"note"`
	TimestampStart time.Time        `json:"start"`
	TimestampEnd   out.EndTimestamp `json:"end"`
	Duration       out.Seconds      `json:"duration"`
}

var Cmd = &cobra.Command{
	Use:     "block [flags] [key | arguments]",
	Aliases: []string{"blocks", "blk", "b"},
	Short:   "zeit block",
	Long:    "View and manage zeit blocks",
	Example: "zeit block 01998b32-7f89-7373-a192-56417e0bc89f",
	Run: func(cmd *cobra.Command, args []string) {
		var blockKey string = ""
		var pargs *argsparser.ParsedArgs
		var blockMap map[string]*block.Block = make(map[string]*block.Block)
		var bvs []BlockView
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), true)
		defer rt.End()

		if len(args) == 1 {
			if strings.Index(args[0], "block:") > -1 {
				blockKey = args[0]
			} else if _, err = uuid.Parse(args[0]); err == nil {
				blockKey = "block:" + args[0]
			}
		}

		if blockKey != "" {
			pargs = new(argsparser.ParsedArgs)
			// Show specific block
			b, err := block.Get(rt.Database, blockKey)
			rt.NilOrDie(err)

			blockMap = make(map[string]*block.Block)
			blockMap[b.GetKey()] = b
		} else {
			pargs, err = argsparser.POP("block", flags, args, rt.Logger)
			rt.NilOrDie(err)

			blockMap, err = block.List(rt.Database)
			rt.NilOrDie(err)
		}

		order := database.GetOrderedKeys(blockMap)
		var newOrder []string
		for _, key := range order {
			timestampStart := pargs.GetTimestampStart()
			timestampEnd := pargs.GetTimestampEnd()

			if timestamp.IsPartiallyWithinTimeframe(
				timestampStart, timestampEnd,
				blockMap[key].TimestampStart, blockMap[key].TimestampEnd) == false {
				continue
			}

			if pargs.ProjectSID != "" {
				if blockMap[key].ProjectSID != pargs.ProjectSID {
					continue
				}
			}

			if pargs.TaskSID != "" {
				if blockMap[key].TaskSID != pargs.TaskSID {
					continue
				}
			}

			var duration time.Duration
			if blockMap[key].TimestampStart.IsZero() == false &&
				blockMap[key].TimestampEnd.IsZero() == false {
				duration = blockMap[key].TimestampEnd.Sub(
					blockMap[key].TimestampStart)
			}

			bvs = append(bvs, BlockView{
				Key:            key,
				ProjectSID:     blockMap[key].ProjectSID,
				TaskSID:        blockMap[key].TaskSID,
				Note:           blockMap[key].Note,
				TimestampStart: blockMap[key].TimestampStart,
				TimestampEnd:   out.EndTimestamp(blockMap[key].TimestampEnd),
				Duration:       out.Seconds(duration),
			})
			newOrder = append(newOrder, key)
		}

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, pargs, bvs, newOrder)
		case FormatCLI:
			outputCLI(rt, pargs, bvs, newOrder)
		case FormatJSON:
			outputJSON(rt, bvs, newOrder)
		}
	},
}

func outputCLI(
	rt *runtime.Runtime,
	pargs *argsparser.ParsedArgs,
	list []BlockView,
	order []string,
) {
	timestampStart := pargs.GetTimestampStart()
	timestampEnd := pargs.GetTimestampEnd()
	if timestampStart.IsZero() == false ||
		timestampEnd.IsZero() == false {

		formatStart := timestampStart.Format(time.DateTime)
		if timestampStart.IsZero() {
			formatStart = "*"
		}

		formatEnd := timestampEnd.Format(time.DateTime)
		if timestampEnd.IsZero() {
			formatEnd = "*"
		}

		rt.Out.Put(out.Opts{Type: out.Info},
			"%s %s %s %s",
			rt.Out.Stylize(
				out.Style{FG: out.ColorSecondary},
				"Timeframe:",
			),
			rt.Out.Stylize(out.Style{FG: out.OutputPrefixes[out.Start].Color},
				"%s",
				formatStart,
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorSecondary},
				"→",
			),
			rt.Out.Stylize(out.Style{FG: out.OutputPrefixes[out.End].Color},
				"%s",
				formatEnd,
			),
		)
	}

	for idx := range order {
		rt.Out.Put(out.Opts{Type: out.Info},
			"%s  %s %s\n  %s %s %s\n  tracked on %s/%s\n  %s",
			rt.Out.Stylize(
				out.Style{FG: out.ColorPrimary},
				"%s", list[idx].Key,
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorSecondary},
				"⭘",
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorWhite},
				"%s", list[idx].Duration,
			),
			rt.Out.Stylize(
				out.Style{FG: out.OutputPrefixes[out.Start].Color},
				"%s%s",
				out.OutputPrefixes[out.Start].Char,
				list[idx].TimestampStart.Format(time.DateTime),
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorSecondary},
				"→",
			),
			rt.Out.Stylize(
				out.Style{FG: out.OutputPrefixes[out.End].Color},
				"%s%s",
				out.OutputPrefixes[out.End].Char,
				list[idx].TimestampEnd,
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorPrimary},
				"%s",
				list[idx].ProjectSID,
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorPrimary},
				"%s",
				list[idx].TaskSID,
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorBrightBlack},
				"%s", block.GetNotePreview(list[idx].Note, 0),
			),
		)

		if idx < len(order)-1 {
			rt.Out.Put(out.Opts{Type: out.Plain}, "")
		}
	}
}

func outputJSON(
	rt *runtime.Runtime,
	list []BlockView,
	order []string,
) {
	prettyJSON, err := json.MarshalIndent(list, "", "  ")
	rt.NilOrDie(err)

	rt.Out.Put(out.Opts{Type: out.Plain}, "%s", string(prettyJSON))
}

func init() {
	Cmd.AddCommand(blockEditCmd.Cmd)

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
}
