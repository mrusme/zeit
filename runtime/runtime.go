package runtime

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/log"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/models/config"
	"github.com/spf13/cobra"
)

var DATABASE_ENV_VAR = "ZEIT_DATABASE"

var (
	Version string
	Commit  string
	Date    string
)

type Build struct {
	Version string
	Commit  string
	Date    string
}

type Runtime struct {
	Build    Build
	Logger   *log.Logger
	Out      *out.Out
	Database *database.Database
	Config   *config.Config
}

func New(lvl slog.Level, oc out.OutputColor) *Runtime {
	var err error

	rt := new(Runtime)

	rt.Build.Version = Version
	rt.Build.Commit = Commit
	rt.Build.Date = Date

	// TODO: Output to file
	rt.Logger = log.New(lvl)

	rt.Out = out.New(oc)

	var dbdir string
	var found bool
	if dbdir, found = os.LookupEnv(DATABASE_ENV_VAR); found == false {
		dbdir, err = xdg.DataFile(path.Join("zeit", "db"))
		rt.Logger.NilOrDie(err, "Could not get database directory: "+
			DATABASE_ENV_VAR+" not set and XDG DataFile reported error!")
		rt.Logger.Debug("Using XDG DataFile directory for database",
			"directory", dbdir,
		)
	} else {
		rt.Logger.Debug("Using "+DATABASE_ENV_VAR+" directory for database",
			"directory", dbdir,
		)
	}

	rt.Database, err = database.New(rt.Logger, dbdir)
	rt.Logger.NilOrDie(err, "Error initializing database")

	rt.Logger.Debug("Loading runtime config ...")
	cfg, err := config.Get(rt.Database)
	if err != nil {
		rt.Logger.NilOrDie(err, "Error loading runtime config")
	}
	rt.Config = cfg
	rt.Logger.Debug("Runtime config loaded")

	return rt
}

func (rt *Runtime) End() {
	rt.Logger.Debug("Ending runtime ...")
	rt.Database.Close()
}

func (rt *Runtime) Exit(code int) {
	rt.End()
	os.Exit(code)
}

func (rt *Runtime) GetUserKey() string {
	return rt.Config.UserKey
}

func (rt *Runtime) GetCommandCall(cmd *cobra.Command) string {
	calledAs := strings.ToLower(cmd.CalledAs())
	if calledAs == "" {
		calledAs = strings.ToLower(cmd.Name())
	}
	return calledAs
}

func (rt *Runtime) GetStringFlag(cmd *cobra.Command, flagname string) string {
	flag, err := cmd.Flags().GetString(flagname)
	rt.Logger.NilOrDie(err, "Could not get "+flagname+" flag")
	return flag
}

func (rt *Runtime) GetIntFlag(cmd *cobra.Command, flagname string) int {
	flag, err := cmd.Flags().GetInt(flagname)
	rt.Logger.NilOrDie(err, "Could not get "+flagname+" flag")
	return flag
}

func (rt *Runtime) GetBoolFlag(cmd *cobra.Command, flagname string) bool {
	flag, err := cmd.Flags().GetBool(flagname)
	rt.Logger.NilOrDie(err, "Could not get "+flagname+" flag")
	return flag
}

func (rt *Runtime) GetDebugFlag(cmd *cobra.Command) bool {
	return rt.GetBoolFlag(cmd, "debug")
}

func (rt *Runtime) GetColorFlag(cmd *cobra.Command) string {
	return rt.GetStringFlag(cmd, "color")
}
