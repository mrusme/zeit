package blockCmd

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	blockEditCmd "github.com/mrusme/zeit/cli/block/edit/cmd"
	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/errs"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/helpers/timestamp"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

const (
	FormatUnspecified = ""
	FormatCLI         = "cli"
	FormatJSON        = "json"
)

var flagFormat string

type BlockView struct {
	Key            string        `json:"key"`
	ProjectSID     string        `json:"project_sid"`
	TaskSID        string        `json:"task_sid"`
	Note           string        `json:"note"`
	TimestampStart time.Time     `json:"start"`
	TimestampEnd   time.Time     `json:"end"`
	Duration       time.Duration `json:"duration"`
}

var Cmd = &cobra.Command{
	Use:     "block [flags] [key | timeframe]",
	Aliases: []string{"blocks", "blk", "b"},
	Short:   "zeit block",
	Long:    "View and manage zeit blocks",
	Example: "zeit block 01998b32-7f89-7373-a192-56417e0bc89f",
	Run: func(cmd *cobra.Command, args []string) {
		var timeframe string = ""
		var tstamp *timestamp.Timestamp
		var blockKey string = ""
		var isBlockKey bool = false
		var dump map[string]*block.Block
		var bvs []BlockView
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		if len(args) == 1 {
			if strings.Index(args[0], "block:") == -1 {
				if _, err = uuid.Parse(args[0]); err == nil {
					blockKey = "block:" + args[0]
					isBlockKey = true
				}
			} else {
				blockKey = args[0]
				isBlockKey = true
			}
		}

		if len(args) > 0 && isBlockKey == false {
			timeframe = strings.Join(args, " ")
			tstamp, err = timestamp.ParsePeriod(timeframe)
			rt.NilOrDie(err)
			if tstamp.IsRange == false {
				rt.NilOrDie(errs.ErrNotATimeframe)
			}

			rt.Out.Put(out.Opts{Type: out.Info},
				"%s %s %s %s",
				rt.Out.Stylize(
					out.Style{FG: out.ColorSecondary},
					"Timeframe:",
				),
				rt.Out.Stylize(out.Style{FG: out.OutputPrefixes[out.Start].Color},
					"%s",
					tstamp.Time.Format(time.DateTime),
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorSecondary},
					"→",
				),
				rt.Out.Stylize(out.Style{FG: out.OutputPrefixes[out.End].Color},
					"%s",
					tstamp.ToTime.Format(time.DateTime),
				),
			)
		}

		if (len(args) == 0 && blockKey == "") ||
			(len(args) > 0 && isBlockKey == false) {
			// List all blocks
			dump, err = block.List(rt.Database)
			rt.NilOrDie(err)
		} else {
			// Show specific block
			tk, err := block.Get(rt.Database, blockKey)
			rt.NilOrDie(err)

			dump = make(map[string]*block.Block)
			dump[tk.GetKey()] = tk
		}

		order := database.GetOrderedKeys(dump)
		var newOrder []string
		for _, key := range order {
			var duration time.Duration

			if isBlockKey == false && tstamp.IsRange == true {
				if (dump[key].TimestampStart.After(tstamp.Time) &&
					dump[key].TimestampStart.Before(tstamp.ToTime)) ||
					(dump[key].TimestampEnd.After(tstamp.Time) &&
						dump[key].TimestampEnd.Before(tstamp.ToTime)) {
				} else {
					continue
				}
			}

			if dump[key].TimestampStart.IsZero() == false &&
				dump[key].TimestampEnd.IsZero() == false {
				duration = dump[key].TimestampEnd.Sub(dump[key].TimestampStart)
			}

			bvs = append(bvs, BlockView{
				Key:            key,
				ProjectSID:     dump[key].ProjectSID,
				TaskSID:        dump[key].TaskSID,
				Note:           dump[key].Note,
				TimestampStart: dump[key].TimestampStart,
				TimestampEnd:   dump[key].TimestampEnd,
				Duration:       duration,
			})
			newOrder = append(newOrder, key)
		}

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, bvs, newOrder)
		case FormatCLI:
			outputCLI(rt, bvs, newOrder)
		case FormatJSON:
			outputJSON(rt, bvs, newOrder)
		}
	},
}

func outputCLI(
	rt *runtime.Runtime,
	list []BlockView,
	order []string,
) {
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
				"%s", list[idx].Duration.Round(time.Second).String(),
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
				list[idx].TimestampEnd.Format(time.DateTime),
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

	rt.Out.Put(out.Opts{Type: out.Plain}, string(prettyJSON))
}

func init() {
	Cmd.AddCommand(blockEditCmd.Cmd)

	Cmd.PersistentFlags().StringVarP(
		&flagFormat,
		"format",
		"f",
		"",
		"Export format (cli, json) (default \"cli\")",
	)
}
