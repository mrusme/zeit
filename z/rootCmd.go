package z

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var database *Database

var begin string
var finish string
var project string
var task string
var notes string

var since string
var until string

var format string
var force bool

var noColors bool

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

	rootCmd.PersistentFlags().BoolVar(&noColors, "no-colors", false, "Do not use colors in output")
}

func initConfig() {
	if noColors == true {
		color.Disable()
	}
}
