package shared

import (
	"github.com/mrusme/zeit/models/project"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

func DynamicArgs(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	var list map[string]*project.Project
	var err error

	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd), true)
	defer rt.End()

	list, err = project.List(rt.Database)
	rt.NilOrDie(err)

	var possibleArgs []string
	for _, pj := range list {
		possibleArgs = append(possibleArgs, pj.SID)
	}

	return rt.GetDynamicSuggestions(toComplete, possibleArgs),
		cobra.ShellCompDirectiveNoFileComp
}
