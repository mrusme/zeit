package z

import (
  "fmt"
  "github.com/spf13/cobra"
  "os"
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

const(
  CharTrack = " ▶"
  CharFinish = " ■"
  CharErase = " ◀"
  CharError = " ▲"
  CharInfo = " ●"
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
}

func initConfig() {
}
