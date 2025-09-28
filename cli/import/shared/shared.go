package shared

import (
	"github.com/spf13/cobra"
)

func DynamicArgs(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
	// defer rt.End()

	return nil,
		cobra.ShellCompDirectiveDefault
}
