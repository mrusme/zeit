package shared

import (
	"slices"
	"strings"

	"github.com/mrusme/zeit/models/task"
	"github.com/mrusme/zeit/runtime"
	"github.com/spf13/cobra"
)

func DynamicArgs(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]cobra.Completion, cobra.ShellCompDirective) {
	var tklist map[string]*task.Task
	var err error

	rt := runtime.New(runtime.GetLogLevel(cmd), runtime.GetOutputColor(cmd))
	defer rt.End()

	var keywords map[string][]string = make(map[string][]string)
	keywords["once"] = []string{"on", "ended", "until", "with"}
	keywords["multi"] = []string{"with"}

	var possibleArgs []string

	for _, keyword := range keywords["once"] {
		if slices.Index(args, keyword) == -1 &&
			LastIsOneOf(args, keywords["once"]) == false &&
			LastIsOneOf(args, keywords["multi"]) == false {
			possibleArgs = append(possibleArgs, keyword)
		}
	}

	if len(args) > 0 {
		last := strings.ToLower(args[len(args)-1])
		switch last {
		case "with":
			if slices.Index(args, "note") == -1 {
				possibleArgs = append(possibleArgs,
					"note",
				)
			}
		case "note":
			possibleArgs = []string{}
		case "on":
			tklist, err = task.List(rt.Database)
			rt.NilOrDie(err)

			for _, tk := range tklist {
				possibleArgs = append(possibleArgs, tk.ProjectSID+"/"+tk.SID)
			}
		default:
			// TODO: Find out how to suggest three words at once, without having
			// cobra escape the spaces with \ characters
			possibleArgs = append(possibleArgs,
				"5 minutes ago",
				"15 minutes ago",
				"1 hour ago",
				"today at 08:00",
				"yesterday at 08:00",
			)
		}
	}

	if len(possibleArgs) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return rt.GetDynamicSuggestions(toComplete, possibleArgs),
		cobra.ShellCompDirectiveNoFileComp
}

func LastIsOneOf(args []string, s []string) bool {
	if len(args) == 0 {
		return false
	}
	last := strings.ToLower(args[len(args)-1])
	return slices.Index(s, last) >= 0
}
