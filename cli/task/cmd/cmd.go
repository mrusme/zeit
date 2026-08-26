package taskCmd

import (
	"encoding/json"
	"strings"
	"time"

	taskEditCmd "xn--gckvb8fzb.com/zeit/cli/task/edit/cmd"
	"xn--gckvb8fzb.com/zeit/cli/task/shared"
	"xn--gckvb8fzb.com/zeit/database"
	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/models/task"
	"xn--gckvb8fzb.com/zeit/runtime"
	"github.com/spf13/cobra"
)

const (
	FormatUnspecified = ""
	FormatCLI         = "cli"
	FormatJSON        = "json"
)

var flagFormat string

type TaskView struct {
	SID         string          `json:"sid"`
	ProjectSID  string          `json:"project_sid"`
	DisplayName string          `json:"display_name"`
	Color       string          `json:"color"`
	Blocks      []TaskBlockView `json:"blocks"`
	TotalBlocks int             `json:"total_blocks"`
	TotalAmount out.Seconds     `json:"total_amount"`
}

type TaskBlockView struct {
	Key            string           `json:"key"`
	Note           string           `json:"note"`
	TimestampStart time.Time        `json:"start"`
	TimestampEnd   out.EndTimestamp `json:"end"`
	Duration       out.Seconds      `json:"duration"`
}

var Cmd = &cobra.Command{
	Use:               "task [flags] [project-sid[/]task-sid]",
	Aliases:           []string{"tasks", "tsk", "tk"},
	Short:             "zeit task",
	Long:              "View and manage zeit tasks",
	Example:           "zeit task myproject mytask",
	Args:              cobra.RangeArgs(0, 2),
	ValidArgsFunction: shared.DynamicArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var dump map[string]*task.Task
		var tkvs []TaskView
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), true)
		defer rt.End()

		var projectSID string
		var taskSID string
		var found bool

		if len(args) == 1 {
			projectSID, taskSID, found = strings.Cut(args[0], "/")
			if found == false {
				projectSID = args[0]
				taskSID = ""
			}
		} else if len(args) == 2 {
			projectSID = args[0]
			taskSID = args[1]
		}

		if taskSID == "" {
			// List all tasks

			if projectSID == "" {
				dump, err = task.List(rt.Database)
			} else {
				dump, err = task.ListForProjectSID(rt.Database, projectSID)
			}
			rt.NilOrDie(err)
		} else {
			// Show specific task
			tk, err := task.GetBySID(rt.Database, projectSID, taskSID)
			rt.NilOrDie(err)

			dump = make(map[string]*task.Task)
			dump[tk.GetKey()] = tk
		}

		order := database.GetOrderedKeys(dump)
		for _, key := range order {
			bs, err := block.ListForProjectTaskSID(rt.Database,
				dump[key].ProjectSID, dump[key].SID)
			rt.NilOrDie(err)

			var bvs []TaskBlockView
			var pTotalBlocks int
			var pTotalAmount time.Duration

			border := database.GetOrderedKeys(bs)
			for _, bkey := range border {

				var duration time.Duration
				if bs[bkey].TimestampStart.IsZero() == false &&
					bs[bkey].TimestampEnd.IsZero() == false {
					duration = bs[bkey].TimestampEnd.Sub(bs[bkey].TimestampStart)
				}

				bvs = append(bvs, TaskBlockView{
					Key:            bkey,
					Note:           bs[bkey].Note,
					TimestampStart: bs[bkey].TimestampStart,
					TimestampEnd:   out.EndTimestamp(bs[bkey].TimestampEnd),
					Duration:       out.Seconds(duration),
				})

				pTotalBlocks += 1
				pTotalAmount += duration
			}

			tkvs = append(tkvs, TaskView{
				SID:         dump[key].SID,
				ProjectSID:  dump[key].ProjectSID,
				DisplayName: dump[key].DisplayName,
				Color:       dump[key].Color,
				Blocks:      bvs,
				TotalBlocks: pTotalBlocks,
				TotalAmount: out.Seconds(pTotalAmount),
			})
		}

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, tkvs, order)
		case FormatCLI:
			outputCLI(rt, tkvs, order)
		case FormatJSON:
			outputJSON(rt, tkvs, order)
		}
	},
}

func outputCLI(
	rt *runtime.Runtime,
	list []TaskView,
	order []string,
) {
	var bcs string = "│"
	var bcc string = "├──"
	var bce string = "└──"

	for idx := range order {
		rt.Out.Put(out.Opts{Type: out.Info},
			"%s%s\n%s  %s %s %s %s",
			rt.Out.Stylize(
				out.Style{BG: out.Color(list[idx].Color), FG: out.ColorBrightWhite, PX: 1},
				"%s", list[idx].DisplayName,
			),
			rt.Out.Stylize(
				out.Style{BG: out.Color(list[idx].Color), FG: out.ColorBlack, PX: 1},
				"[%s/%s]", list[idx].ProjectSID, list[idx].SID,
			),
			rt.Out.Stylize(
				out.Style{FG: out.OutputPrefixes[out.Info].Color},
				"%s", bcs,
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorSecondary},
				"⮻",
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorWhite},
				"%-6d", list[idx].TotalBlocks,
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorSecondary},
				"⭘",
			),
			rt.Out.Stylize(
				out.Style{FG: out.ColorWhite},
				"%-12s", list[idx].TotalAmount,
			),
		)

		for jdx := range list[idx].Blocks {
			barchar := bcc
			barcharsub := bcs
			if jdx == len(list[idx].Blocks)-1 {
				barchar = bce
				barcharsub = " "
			}

			rt.Out.Put(out.Opts{Type: out.Plain},
				"%s %s  %s %s\n%s   %s %s %s\n%s   %s",
				rt.Out.Stylize(
					out.Style{FG: out.OutputPrefixes[out.Info].Color},
					"%s", barchar,
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorPrimary},
					"%s", list[idx].Blocks[jdx].Key,
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorSecondary},
					"⭘",
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorWhite},
					"%s", list[idx].Blocks[jdx].Duration,
				),
				rt.Out.Stylize(
					out.Style{FG: out.OutputPrefixes[out.Info].Color},
					"%s", barcharsub,
				),
				rt.Out.Stylize(
					out.Style{FG: out.OutputPrefixes[out.Start].Color},
					"%s%s",
					out.OutputPrefixes[out.Start].Char,
					list[idx].Blocks[jdx].TimestampStart.Format(time.DateTime),
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorSecondary},
					"→",
				),
				rt.Out.Stylize(
					out.Style{FG: out.OutputPrefixes[out.End].Color},
					"%s%s",
					out.OutputPrefixes[out.End].Char,
					list[idx].Blocks[jdx].TimestampEnd,
				),
				rt.Out.Stylize(
					out.Style{FG: out.OutputPrefixes[out.Info].Color},
					"%s", barcharsub,
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorBrightBlack},
					"%s", block.GetNotePreview(list[idx].Blocks[jdx].Note, 0),
				),
			)
		}

		if idx < len(order)-1 {
			rt.Out.Put(out.Opts{Type: out.Plain}, "")
		}
	}
}

func outputJSON(
	rt *runtime.Runtime,
	list []TaskView,
	order []string,
) {
	prettyJSON, err := json.MarshalIndent(list, "", "  ")
	rt.NilOrDie(err)

	rt.Out.Put(out.Opts{Type: out.Plain}, "%s", string(prettyJSON))
}

func init() {
	Cmd.AddCommand(taskEditCmd.Cmd)

	Cmd.PersistentFlags().StringVarP(
		&flagFormat,
		"format",
		"f",
		"",
		"Output format (cli, json) (default \"cli\")",
	)
}
