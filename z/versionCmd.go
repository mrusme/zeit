package z

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VERSION string

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display what Zeit it is",
	Long:  `The version of Zeit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("zeit", VERSION)
	},
}
