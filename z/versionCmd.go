package z

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display what Zeit it is",
	Long:  `The version of Zeit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("zeit %s, commit %s, built at %s\n", version, commit, date)
	},
}
