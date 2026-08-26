package blockFixCmd

import (
	"bufio"
	"os"
	"strings"
	"time"

	"xn--gckvb8fzb.com/zeit/errs"
	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/helpers/timestamp"
	"xn--gckvb8fzb.com/zeit/models/block"
	"xn--gckvb8fzb.com/zeit/runtime"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "fix",
	Aliases: []string{},
	Short:   "zeit block fix",
	Long:    "End blocks that were left without an end timestamp",
	Example: "zeit block fix",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), false)
		defer rt.End()

		blockMap, err := block.List(rt.Database)
		rt.NilOrDie(err)

		found, _, running, err := block.GetActive(rt.Database)
		rt.NilOrDie(err)

		var activeBlockKey string
		if found == true {
			activeBlockKey = running.GetKey()
		}

		unfinished := block.ListUnfinished(blockMap, activeBlockKey)
		if len(unfinished) == 0 {
			rt.Out.Put(out.Opts{Type: out.Ok}, "No blocks left to fix")
			return
		}

		rt.Out.Put(out.Opts{Type: out.Info},
			"%d %s without an end timestamp",
			len(unfinished),
			pluralize(len(unfinished), "block", "blocks"),
		)
		rt.Out.Put(out.Opts{Type: out.Plain}, "")

		fixed := fixBlocks(rt, bufio.NewReader(os.Stdin), unfinished)

		rt.Out.Put(out.Opts{Type: out.Ok},
			"Fixed %d of %d %s",
			fixed,
			len(unfinished),
			pluralize(len(unfinished), "block", "blocks"),
		)
	},
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}

func fixBlocks(
	rt *runtime.Runtime,
	reader *bufio.Reader,
	unfinished []block.Unfinished,
) int {
	var fixed int

	for _, u := range unfinished {
		outputBlock(rt, u.Block)

		for {
			end, err := askForEnd(rt, reader, u.RecommendedEnd)
			if err != nil {
				rt.Out.Put(out.Opts{Type: out.Plain}, "")
				return fixed
			}

			if end.IsZero() == true {
				rt.Out.Put(out.Opts{Type: out.Pause}, "Left unchanged")
				rt.Out.Put(out.Opts{Type: out.Plain}, "")
				break
			}

			u.Block.TimestampEnd = end

			if err = block.Set(rt.Database, u.Block); err != nil {
				rt.Out.Put(out.Opts{Type: out.Error}, "%s", err.Error())
				continue
			}

			rt.Out.Put(out.Opts{Type: out.End},
				"Ended at %s",
				rt.Out.FG(out.ColorPrimary, "%s", end.Format(time.DateTime)),
			)
			rt.Out.Put(out.Opts{Type: out.Plain}, "")

			fixed++
			break
		}
	}

	return fixed
}

func askForEnd(
	rt *runtime.Runtime,
	reader *bufio.Reader,
	recommended time.Time,
) (time.Time, error) {
	for {
		if recommended.IsZero() == true {
			rt.Out.Put(out.Opts{Type: out.Plain, NoNL: true},
				"  %s ",
				rt.Out.FG(out.ColorSecondary,
					"End at (empty to leave it unchanged):"),
			)
		} else {
			rt.Out.Put(out.Opts{Type: out.Plain, NoNL: true},
				"  %s %s ",
				rt.Out.FG(out.ColorSecondary, "End at (empty to accept):"),
				rt.Out.FG(out.ColorPrimary, "%s", recommended.Format(time.DateTime)),
			)
		}

		answer, err := reader.ReadString('\n')
		if err != nil && answer == "" {
			return time.Time{}, err
		}

		answer = strings.TrimSpace(answer)
		if answer == "" {
			return recommended, nil
		}

		ts, err := timestamp.Parse(answer)
		if err != nil {
			rt.Out.Put(out.Opts{Type: out.Error}, "%s", (&errs.ErrParsingTimestamp{
				Message:   err.Error(),
				Timestamp: answer,
			}).Error())
			continue
		}

		return ts.Time, nil
	}
}

func outputBlock(rt *runtime.Runtime, b *block.Block) {
	rt.Out.Put(out.Opts{Type: out.Info},
		"%s\n  %s %s %s\n  tracked on %s\n  %s",
		rt.Out.Stylize(
			out.Style{FG: out.ColorPrimary},
			"%s", b.GetKey(),
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
			out.Style{FG: out.ColorPrimary},
			"%s/%s", b.ProjectSID, b.TaskSID,
		),
		rt.Out.Stylize(
			out.Style{FG: out.ColorBrightBlack},
			"%s", block.GetNotePreview(b.Note, 0),
		),
	)
}
