package z

import (
	"fmt"
	"sort"
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type reportEntry struct {
	Date     string
	Project  string
	Task     string
	Duration float64
	Notes    string
	Running  bool
}

type reportLine struct {
	Duration float64
	Notes    []string
	Running  bool
}

var (
	weeklyFlag  bool
	monthlyFlag bool
	notesFlag   bool
	noTasksFlag bool
)
var dailyReport map[string]map[string]map[string]reportLine

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "report times an day / project / task level",
	Long:  "Reporting summaries on daily, project, task level for a given range",
	Run: func(cmd *cobra.Command, args []string) {
		if since == "" && until == "" && listRange == "" {
			listRange = viper.GetString("report.default")
		}

		dailyReport = make(map[string]map[string]map[string]reportLine)

		if weeklyFlag {
			viper.Set("report.weeklySum", true)
		}
		if monthlyFlag {
			viper.Set("report.monthlySum", true)
		}
		if notesFlag {
			viper.Set("report.notes", true)
		}
		if noTasksFlag {
			viper.Set("report.no-tasks", true)
		}

		filteredEntries := listEntries()
		sinceTime, untilTime := ParseSinceUntil(since, until, listRange)
		if listRange != "" {
			fmt.Println("Reporting for Timerange:", listRange, "/", sinceTime.Format(DateFormat), "-", untilTime.Format(DateFormat))
		}
		var reportEntries []reportEntry
		for _, fe := range filteredEntries {
			var entryDuration float64
			running := false
			if fe.Finish.IsZero() {
				entryDuration = time.Duration(time.Since(fe.Begin)).Seconds()
				running = true
			} else {
				entryDuration = time.Duration(fe.Finish.Sub(fe.Begin)).Seconds()
			}
			dateString := fe.Begin.Format(DateFormat)
			reportEntries = append(reportEntries, reportEntry{dateString, fe.Project, fe.Task, entryDuration, fe.Notes, running})
		}

		for _, re := range reportEntries {
			dailyReporting(re)
		}

		output()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVar(&since, "since", "", "Date/time to start the list from")
	reportCmd.Flags().StringVar(&until, "until", "", "Date/time to list until")
	reportCmd.Flags().StringVar(&listRange, "range", "", "shortcut to set since/until for a given range (today, yesterday, thisWeek, lastWeek, thisMonth, lastMonth)")
	reportCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be listed")
	reportCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be listed")
	reportCmd.PersistentFlags().BoolVar(&weeklyFlag, "weekly", false, "Print summary of weekly hours")
	reportCmd.PersistentFlags().BoolVar(&monthlyFlag, "monthly", false, "Print summary of monthly hours")
	reportCmd.PersistentFlags().BoolVar(&notesFlag, "notes", false, "Print notes for the task")
	reportCmd.PersistentFlags().BoolVar(&noTasksFlag, "no-tasks", false, "Print only summary bot no task details")

	flagName := "task"
	reportCmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		user := GetCurrentUser()
		entries, _ := database.ListEntries(user)
		_, tasks := listProjectsAndTasks(entries)
		return tasks, cobra.ShellCompDirectiveDefault
	})
}

func dailyReporting(re reportEntry) {
	_, ok := dailyReport[re.Date]
	if !ok {
		dailyReport[re.Date] = make(map[string]map[string]reportLine)
		dailyReport[re.Date][re.Project] = make(map[string]reportLine)
		dailyReport[re.Date][re.Project][re.Task] = reportLine{Duration: re.Duration, Notes: []string{re.Notes}, Running: re.Running}
		return
	}

	_, ok = dailyReport[re.Date][re.Project]
	if !ok {
		dailyReport[re.Date][re.Project] = make(map[string]reportLine)
		dailyReport[re.Date][re.Project][re.Task] = reportLine{Duration: re.Duration, Notes: []string{re.Notes}, Running: re.Running}
		return
	}

	_, ok = dailyReport[re.Date][re.Project][re.Task]
	if !ok {
		dailyReport[re.Date][re.Project][re.Task] = reportLine{Duration: re.Duration, Notes: []string{re.Notes}, Running: re.Running}
		return
	}

	workEntry := dailyReport[re.Date][re.Project][re.Task]
	if workEntry.Running || re.Running {
		workEntry.Running = true
	}
	workEntry.Duration += re.Duration
	workEntry.Notes = append(workEntry.Notes, re.Notes)
	dailyReport[re.Date][re.Project][re.Task] = workEntry
}

func output() {
	lastWeek := ""
	weekSum := 0.0
	lastMonth := ""
	monthSum := 0.0
	for _, dateKey := range dialyKeys() {
		dailySum := 0.0
		fmt.Println("  ")
		for _, projectKey := range projectKeys(dateKey) {
			projectSum := 0.0
			for _, taskKey := range taskKeys(dateKey, projectKey) {
				t, _ := time.Parse("2006-01-02", dateKey)
				year, week := t.ISOWeek()
				thisWeek := fmt.Sprintf("%04d-%02d", year, week)
				if lastWeek != "" && lastWeek != thisWeek {
					if viper.GetBool("report.weeklySum") {
						color.FgGray.Println("  Week: ", lastWeek, ":", fmtDuration(time.Duration(weekSum*float64(time.Second))), "\n-------------------\n")
					}
					lastWeek = thisWeek
					weekSum = 0.0
				}
				if lastWeek == "" {
					lastWeek = thisWeek
				}

				month := t.Month()
				thisMonth := fmt.Sprintf("%04d-%02d", year, month)
				if lastMonth != "" && lastMonth != thisMonth {
					if viper.GetBool("report.monthlySum") {
						color.FgGray.Println(" Month: ", lastMonth, ":", fmtDuration(time.Duration(monthSum*float64(time.Second))), "\n=====================\n")
					}
					lastMonth = thisMonth
					monthSum = 0.0
				}
				if lastMonth == "" {
					lastMonth = thisMonth
				}

				if !viper.GetBool("report.no-tasks") {
					color.FgLightWhite.Print("          ", fmtDuration(time.Duration(dailyReport[dateKey][projectKey][taskKey].Duration*float64(time.Second))), " ", taskKey)
					if dailyReport[dateKey][projectKey][taskKey].Running {
						color.FgLightYellow.Println(" (running)")
					} else {
						fmt.Println()
					}
					if viper.GetBool("report.notes") {
						for _, note := range dailyReport[dateKey][projectKey][taskKey].Notes[1:] {
							if len(note) > 0 {
								color.FgLightBlue.Println("                    ", note)
							}
						}
					}
				}
				projectSum += dailyReport[dateKey][projectKey][taskKey].Duration
				dailySum += dailyReport[dateKey][projectKey][taskKey].Duration
				weekSum += dailyReport[dateKey][projectKey][taskKey].Duration
				monthSum += dailyReport[dateKey][projectKey][taskKey].Duration
			}
			fmt.Println("       ", projectKey, ":", fmtDuration(time.Duration(projectSum*float64(time.Second))))
		}
		fmt.Println("     ", dateKey, ":", fmtDuration(time.Duration(dailySum*float64(time.Second))))
	}
	if viper.GetBool("report.weeklySum") {
		fmt.Println("\n  Week: ", lastWeek, ":", fmtDuration(time.Duration(weekSum*float64(time.Second))))
	}
	if viper.GetBool("report.monthlySUm") {
		fmt.Println("\n Month: ", lastMonth, ":", fmtDuration(time.Duration(monthSum*float64(time.Second))))
	}
}

func dialyKeys() []string {
	keys := make([]string, 0, len(dailyReport))

	for k := range dailyReport {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func projectKeys(daily string) []string {
	keys := make([]string, 0, len(dailyReport[daily]))

	for k := range dailyReport[daily] {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func taskKeys(daily string, project string) []string {
	keys := make([]string, 0, len(dailyReport[daily][project]))

	for k := range dailyReport[daily][project] {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
