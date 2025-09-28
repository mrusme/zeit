package shared

import (
	"github.com/mrusme/zeit/models/task"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

func DynamicArgs(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	var list map[string]*task.Task
	var err error

	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
	defer rt.End()

	list, err = task.List(rt.Database)
	rt.NilOrDie(err)

	var possibleArgs []string
	for _, tk := range list {
		possibleArgs = append(possibleArgs, tk.ProjectSID+"/"+tk.SID)
	}

	return rt.GetDynamicSuggestions(toComplete, possibleArgs),
		cobra.ShellCompDirectiveNoFileComp
}
