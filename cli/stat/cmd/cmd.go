package statCmd

import (
	"encoding/json"
	"maps"
	"slices"
	"time"

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

		pargs, err = argsparser.POP("stat", flags, args, rt.Logger)
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

		tracking, _, ab, err := block.GetActive(rt.Database)
		rt.NilOrDie(err)

		var activeBlockKey *string
		if tracking {
			k := ab.GetKey()
			activeBlockKey = &k
		}

		aggregatedStats, err := aggregateDurations(
			bs, period, timestampStart, timestampEnd, activeBlockKey)
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

func aggregateDurations(
	bs []*block.Block,
	timeframe string,
	timeframeStart time.Time,
	timeframeEnd time.Time,
	activeBlockKey *string,
) (map[string]map[string]map[string]time.Duration, error) {
	period, err := timestamp.GetPeriod(timeframe)
	if err != nil {
		return nil, err
	}

	aggregatedStats := make(map[string]map[string]map[string]time.Duration)

	aggregatedStats["*"] = make(map[string]map[string]time.Duration)
	aggregatedStats["*"]["*"] = make(map[string]time.Duration)
	aggregatedStats["*"]["*"]["*"] = 0

	for _, b := range bs {
		blockEnd := b.TimestampEnd
		if activeBlockKey != nil && b.GetKey() == *activeBlockKey {
			blockEnd = time.Now()
		}

		start, end := timestamp.ClipToTimeframe(
			timeframeStart, timeframeEnd, b.TimestampStart, blockEnd)

		for bucket := period.Start(start); bucket.Before(end); {
			next := period.Next(bucket)

			duration := timestamp.DurationWithinTimeframe(bucket, next, start, end)
			if duration > 0 {
				addDuration(aggregatedStats, b.ProjectSID, b.TaskSID,
					period.Key(bucket), duration)
			}

			bucket = next
		}
	}

	return aggregatedStats, nil
}

func addDuration(
	aggregatedStats map[string]map[string]map[string]time.Duration,
	projectSID string,
	taskSID string,
	key string,
	duration time.Duration,
) {
	if _, ok := aggregatedStats[projectSID]; !ok {
		aggregatedStats[projectSID] = make(map[string]map[string]time.Duration)
	}
	if _, ok := aggregatedStats[projectSID][taskSID]; !ok {
		aggregatedStats[projectSID][taskSID] = make(map[string]time.Duration)
	}

	aggregatedStats[projectSID][taskSID][key] += duration
	aggregatedStats["*"]["*"]["*"] += duration
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

	for _, projectSID := range slices.Sorted(maps.Keys(aggregatedStats)) {
		if projectSID == "*" {
			continue
		}

		taskStats := aggregatedStats[projectSID]

		rt.Out.Put(out.Opts{Type: out.Info},
			"%s",
			projectSID,
		)
		for _, taskSID := range slices.Sorted(maps.Keys(taskStats)) {
			timeframeStats := taskStats[taskSID]

			rt.Out.Put(out.Opts{Type: out.Plain},
				"    %s",
				taskSID,
			)
			for _, timeframe := range slices.Sorted(maps.Keys(timeframeStats)) {
				rt.Out.Put(out.Opts{Type: out.Plain},
					"      %s: %v",
					timeframe,
					timeframeStats[timeframe].Truncate(time.Second),
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
