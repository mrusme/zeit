package statCmd

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/errs"
	"github.com/mrusme/zeit/helpers/argsparser"
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

var (
	flagFormat string
	flags      *argsparser.ParsedArgs
)

// type StatView struct {
// 	ProjectSID     string        `json:"project_sid"`
// 	TaskSID        string        `json:"task_sid"`
// 	TimestampStart time.Time     `json:"start"`
// 	TimestampEnd   time.Time     `json:"end"`
// 	TotalBlocks    int           `json:"total_blocks"`
// 	TotalAmount    time.Duration `json:"total_amount"`
// 	SubUnit        []StatView    `json:"sub_unit"`
// }

var Cmd = &cobra.Command{
	Use:     "stat [flags] [arguments]",
	Aliases: []string{"stats", "stt"},
	Short:   "zeit stat",
	Long:    "View zeit statistics",
	Example: "zeit stat on myproject/mytask this week",
	Run: func(cmd *cobra.Command, args []string) {
		var pargs *argsparser.ParsedArgs
		var blockMap map[string]*block.Block = make(map[string]*block.Block)
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), true)
		defer rt.End()

		pargs, err = argsparser.POP("block", flags, args, rt.Logger)
		rt.NilOrDie(err)

		blockMap, err = block.List(rt.Database)
		rt.NilOrDie(err)

		timestampStart := pargs.GetTimestampStart()
		timestampEnd := pargs.GetTimestampEnd()

		var bs []*block.Block
		order := database.GetOrderedKeys(blockMap)
		var newOrder []string
		for _, key := range order {

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

			bs = append(bs, blockMap[key])

			newOrder = append(newOrder, key)
		}

		diff := timestampEnd.Sub(timestampStart)
		period := "day"
		if diff >= (90 * 24 * time.Hour) {
			period = "month"
		} else if diff >= (14 * 24 * time.Hour) {
			period = "week"
		}

		_, aggregatedStats, err := aggregateDurations(bs, period)
		rt.NilOrDie(err)

		switch flagFormat {
		case FormatUnspecified:
			outputCLI(rt, pargs, aggregatedStats, newOrder)
		case FormatCLI:
			outputCLI(rt, pargs, aggregatedStats, newOrder)
		case FormatJSON:
			outputJSON(rt, aggregatedStats, newOrder)
		}
	},
}

func getDayKey(timestamp time.Time) string {
	return timestamp.Format("2006-01-02") // YYYY-MM-DD format
}

func getWeekKey(timestamp time.Time) string {
	_, week := timestamp.ISOWeek()
	year := timestamp.Year()
	return fmt.Sprintf("%d-W%d", year, week)
}

func getMonthKey(timestamp time.Time) string {
	return timestamp.Format("2006-01") // YYYY-MM format
}

func aggregateDurations(
	bs []*block.Block,
	timeframe string,
) (
	map[string]map[string][]string,
	map[string]map[string]map[string]time.Duration,
	error,
) {
	aggregatedStats := make(map[string]map[string]map[string]time.Duration)
	tfIndex := make(map[string]map[string][]string)

	aggregatedStats["*"] = make(map[string]map[string]time.Duration)
	aggregatedStats["*"]["*"] = make(map[string]time.Duration)
	aggregatedStats["*"]["*"]["*"] = 0

	for _, b := range bs {
		duration := b.TimestampEnd.Sub(b.TimestampStart)

		var key string
		switch timeframe {
		case "day":
			key = getDayKey(b.TimestampStart)
		case "week":
			key = getWeekKey(b.TimestampStart)
		case "month":
			key = getMonthKey(b.TimestampStart)
		default:
			return nil, nil, errs.ErrNotATimeframe
		}

		if _, ok := tfIndex[key]; !ok {
			tfIndex[key] = make(map[string][]string)
		}
		if _, ok := tfIndex[key][b.ProjectSID]; !ok {
			tfIndex[key][b.ProjectSID] = make([]string, 0)
		}

		if _, ok := aggregatedStats[b.ProjectSID]; !ok {
			aggregatedStats[b.ProjectSID] = make(map[string]map[string]time.Duration)
		}

		if _, ok := aggregatedStats[b.ProjectSID][b.TaskSID]; !ok {
			aggregatedStats[b.ProjectSID][b.TaskSID] = make(map[string]time.Duration)
		}

		aggregatedStats[b.ProjectSID][b.TaskSID][key] += duration
		aggregatedStats["*"]["*"]["*"] += duration
		if i := slices.Index(tfIndex[key][b.ProjectSID], b.TaskSID); i == -1 {
			tfIndex[key][b.ProjectSID] = append(tfIndex[key][b.ProjectSID], b.TaskSID)
		}
	}

	return tfIndex, aggregatedStats, nil
}

func outputCLI(
	rt *runtime.Runtime,
	pargs *argsparser.ParsedArgs,
	aggregatedStats map[string]map[string]map[string]time.Duration,
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

	for projectSID, taskStats := range aggregatedStats {
		if projectSID == "*" {
			continue
		}
		rt.Out.Put(out.Opts{Type: out.Info},
			"%s",
			projectSID,
		)
		for taskSID, timeframeStats := range taskStats {
			rt.Out.Put(out.Opts{Type: out.Plain},
				"    %s",
				taskSID,
			)
			for timeframe, totalDuration := range timeframeStats {
				rt.Out.Put(out.Opts{Type: out.Plain},
					"      %s: %v",
					timeframe,
					totalDuration.Truncate(time.Second),
				)
			}
		}
	}

	rt.Out.Put(out.Opts{Type: out.Info},
		"%s %s",
		rt.Out.Stylize(
			out.Style{FG: out.ColorSecondary},
			"Total:",
		),
		rt.Out.Stylize(out.Style{FG: out.OutputPrefixes[out.Start].Color},
			"%v",
			aggregatedStats["*"]["*"]["*"].Truncate(time.Second),
		),
	)
}

func outputJSON(
	rt *runtime.Runtime,
	aggregatedStats map[string]map[string]map[string]time.Duration,
	order []string,
) {
	prettyJSON, err := json.MarshalIndent(aggregatedStats, "", "  ")
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
}
