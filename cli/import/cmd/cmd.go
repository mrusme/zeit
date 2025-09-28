package importCmd

import (
	"errors"
	"strings"

	"github.com/mrusme/zeit/cli/import/shared"
	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/importer"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/activeblock"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/models/config"
	"github.com/mrusme/zeit/models/project"
	"github.com/mrusme/zeit/models/task"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

var flagFormat string

var Cmd = &cobra.Command{
	Use:               "import [flags] file",
	Aliases:           []string{"imp", "im", "i"},
	Short:             "zeit import",
	Long:              "Import data into the zeit database from various formats",
	Example:           "zeit import -f v0 ./zeit-v0-export.json",
	ValidArgsFunction: shared.DynamicArgs,
	Args:              cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var im *importer.Importer
		var err error

		rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
		defer rt.End()

		flagFormat = strings.ToLower(flagFormat)

		im, err = importer.New(importer.ImportFileType(flagFormat), args[0])
		rt.NilOrDie(err)

		err = im.Import(ImportCallback, rt)
		rt.NilOrDie(err)

		rt.Out.Put(out.Opts{Type: out.Ok}, "Import finished")
	},
}

func ImportCallback(
	entry database.Model,
	err error,
	v ...any,
) error {
	var rt *runtime.Runtime
	var ok bool

	if rt, ok = v[0].(*runtime.Runtime); ok == false {
		return errors.New("CRITICAL: Could not retrieve runtime!")
	}

	if err != nil {
		rt.Out.Put(out.Opts{Type: out.Error},
			"An error occurred during import: %s",
			err.Error(),
		)
		// TODO: Shall we ask for user confirmation to continue?
		return nil
	}

	switch model := entry.(type) {
	case *activeblock.ActiveBlock:
		if err = activeblock.Set(rt.Database, model); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error},
				"ActiveBlock could not be stored: %s",
				err.Error(),
			)
			rt.Logger.Debug(err.Error(), "activeblock", model)
			// TODO: Shall we ask for user confirmation to continue?
		}
	case *block.Block:
		model.OwnerKey = rt.Config.UserKey
		if err = block.Set(rt.Database, model); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error},
				"Block with key %s for projectSID/taskSID '%s/%s' could not be stored: %s",
				model.GetKey(),
				model.ProjectSID,
				model.TaskSID,
				err.Error(),
			)
			rt.Logger.Debug(err.Error(), "block", model)
			// TODO: Shall we ask for user confirmation to continue?
		}
	case *config.Config:
		if err = config.Set(rt.Database, model); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error},
				"Config could not be stored: %s",
				err.Error(),
			)
			rt.Logger.Debug(err.Error(), "config", model)
			// TODO: Shall we ask for user confirmation to continue?
		}
	case *project.Project:
		model.OwnerKey = rt.Config.UserKey
		if err = project.Set(rt.Database, model); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error},
				"Project with SID %s could not be stored: %s",
				model.SID,
				err.Error(),
			)
			rt.Logger.Debug(err.Error(), "project", model)
			// TODO: Shall we ask for user confirmation to continue?
		}
	case *task.Task:
		model.OwnerKey = rt.Config.UserKey
		if err = task.Set(rt.Database, model); err != nil {
			rt.Out.Put(out.Opts{Type: out.Error},
				"Task with SID %s could not be stored: %s",
				model.SID,
				err.Error(),
			)
			rt.Logger.Debug(err.Error(), "task", model)
			// TODO: Shall we ask for user confirmation to continue?
		}
	default:
		rt.Out.Put(out.Opts{Type: out.Error},
			"Retrieved unknown data type from importer, wtf.",
		)
	}

	return nil
}

func init() {
	Cmd.PersistentFlags().StringVarP(
		&flagFormat,
		"format",
		"f",
		"v1",
		"Import format (v0, v1)",
	)
}
