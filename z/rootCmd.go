package z

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var database *Database

var (
	begin        string
	finish       string
	switchString string
	project      string
	task         string
	notes        string
)

var (
	since     string
	until     string
	listRange string
)

var (
	format string
	force  bool
)

var (
	noColors bool
	debug    bool
	cfgFile  string
)

const (
	CharTrack  = " ▶"
	CharFinish = " ■"
	CharErase  = " ◀"
	CharError  = " ▲"
	CharInfo   = " ●"
	CharMore   = " ◆"
)

var rootCmd = &cobra.Command{
	Use:   "zeit",
	Short: "Command line Zeiterfassung",
	Long:  `A command line time tracker.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/zeit.[yaml|toml")

	rootCmd.PersistentFlags().BoolVar(&noColors, FlagNoColors, false, "Do not use colors in output")
	viper.BindPFlag(FlagNoColors, rootCmd.PersistentFlags().Lookup(FlagNoColors))

	rootCmd.PersistentFlags().BoolVarP(&debug, FlagDebug, "d", false, "Display debugging output in the console. (default: false)")
	viper.BindPFlag(FlagDebug, rootCmd.PersistentFlags().Lookup(FlagDebug))
}

func initConfig() {
	if noColors == true {
		color.Disable()
	}

	viper.SetEnvPrefix("zeit")
	viper.BindEnv("db")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath("$XDG_CONFIG_HOME")
		viper.AddConfigPath("$XDG_CONFIG_HOME/zeit")
		viper.AddConfigPath(home + "/.config")
		viper.AddConfigPath(home + "/.config/zeit")
		viper.SetConfigName("zeit")
	}

	if err := viper.ReadInConfig(); err != nil {
		// Set default values for parameters
		viper.Set("debug", false)
	}

	if viper.GetBool("debug") {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		fmt.Fprintln(os.Stderr, "Using Database file:", viper.GetString("db"))
	}

	var err error
	database, err = InitDatabase()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
}
