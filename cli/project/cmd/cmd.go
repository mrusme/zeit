package projectCmd

import (
	"encoding/json"
	"time"

	projectEditCmd "github.com/mrusme/zeit/cli/project/edit/cmd"
	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/block"
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

var flagFormat string

type ProjectView struct {
	SID         string            `json:"sid"`
	DisplayName string            `json:"display_name"`
	Color       string            `json:"color"`
	Tasks       []ProjectTaskView `json:"tasks"`
	TotalBlocks int               `json:"total_blocks"`
	TotalAmount time.Duration     `json:"total_amount"`
}

type ProjectTaskView struct {
	SID         string        `json:"sid"`
	DisplayName string        `json:"display_name"`
	Color       string        `json:"color"`
	TotalBlocks int           `json:"total_blocks"`
	TotalAmount time.Duration `json:"total_amount"`
}

var Cmd = &cobra.Command{
	Use:     "project [flags] [sid]",
	Aliases: []string{"projects", "proj", "prj", "pj"},
	Short:   "zeit project",
	Long:    "View and manage zeit projects",
	Example: "zeit project myproject",
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var dump map[string]*project.Project
		var pjvs []ProjectView
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		if len(args) == 0 {
			// List all projects

			dump, err = project.List(rt.Database)
			rt.NilOrDie(err)

		} else {
			// Show specific project
			pj, err := project.GetBySID(rt.Database, args[0])
			rt.NilOrDie(err)

			dump = make(map[string]*project.Project)
			dump[pj.GetKey()] = pj
		}

		order := database.GetOrderedKeys(dump)
		for _, key := range order {
			tks, err := task.ListForProjectSID(rt.Database, dump[key].SID)
			rt.NilOrDie(err)

			var tkvs []ProjectTaskView
			var pTotalBlocks int
			var pTotalAmount time.Duration

			torder := database.GetOrderedKeys(tks)
			for _, tkey := range torder {
				bs, err := block.ListForProjectTaskSID(rt.Database, dump[key].SID, tks[tkey].SID)
				rt.NilOrDie(err)

				var totalAmount time.Duration
				for bkey := range bs {
					if bs[bkey].TimestampStart.IsZero() == false &&
						bs[bkey].TimestampEnd.IsZero() == false {
						duration := bs[bkey].TimestampEnd.Sub(bs[bkey].TimestampStart)
						totalAmount += duration
					}
				}

				totalBlocks := len(bs)

				tkvs = append(tkvs, ProjectTaskView{
					SID:         tks[tkey].SID,
					DisplayName: tks[tkey].DisplayName,
					Color:       tks[tkey].Color,
					TotalBlocks: totalBlocks,
					TotalAmount: totalAmount,
				})

				pTotalBlocks += totalBlocks
				pTotalAmount += totalAmount
			}

			pjvs = append(pjvs, ProjectView{
				SID:         dump[key].SID,
				DisplayName: dump[key].DisplayName,
				Color:       dump[key].Color,
				Tasks:       tkvs,
				TotalBlocks: pTotalBlocks,
				TotalAmount: pTotalAmount,
			})
		}

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, pjvs, order)
		case FormatCLI:
			outputCLI(rt, pjvs, order)
		case FormatJSON:
			outputJSON(rt, pjvs, order)
		}
	},
}

func outputCLI(
	rt *runtime.Runtime,
	list []ProjectView,
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
				"[%s]", list[idx].SID,
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
				"%-12s", list[idx].TotalAmount.Round(time.Second).String(),
			),
		)

		for jdx := range list[idx].Tasks {
			barchar := bcc
			barcharsub := bcs
			if jdx == len(list[idx].Tasks)-1 {
				barchar = bce
				barcharsub = " "
			}
			rt.Out.Put(out.Opts{Type: out.Plain},
				"%s %s %s\n%s   %s %s %s %s",
				rt.Out.Stylize(
					out.Style{FG: out.OutputPrefixes[out.Info].Color},
					"%s", barchar,
				),
				rt.Out.Stylize(
					out.Style{FG: out.Color(list[idx].Tasks[jdx].Color)},
					"%s", list[idx].Tasks[jdx].DisplayName,
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorSecondary},
					"[%s]", list[idx].Tasks[jdx].SID,
				),
				rt.Out.Stylize(
					out.Style{FG: out.OutputPrefixes[out.Info].Color},
					"%s", barcharsub,
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorSecondary},
					"⮻",
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorWhite},
					"%-6d", list[idx].Tasks[jdx].TotalBlocks,
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorSecondary},
					"⭘",
				),
				rt.Out.Stylize(
					out.Style{FG: out.ColorWhite},
					"%-12s", list[idx].Tasks[jdx].TotalAmount.Round(time.Second).String(),
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
	list []ProjectView,
	order []string,
) {
	prettyJSON, err := json.MarshalIndent(list, "", "  ")
	rt.NilOrDie(err)

	rt.Out.Put(out.Opts{Type: out.Plain}, string(prettyJSON))
}

func init() {
	Cmd.AddCommand(projectEditCmd.Cmd)

	Cmd.PersistentFlags().StringVarP(
		&flagFormat,
		"format",
		"f",
		"",
		"Export format (cli, json) (default \"cli\")",
	)
}
