package z

import (
  "fmt"
  "os"
  "time"

  "github.com/spf13/viper"
)

type Task struct {
  Name          string      `json:"name,omitempty"`
  GitRepository string      `json:"gitRepository,omitempty"`
}

func listEntries() []Entry {
  user := GetCurrentUser()

  entries, err := database.ListEntries(user)
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }

  sinceTime, untilTime := ParseSinceUntil(since, until, listRange)

  var filteredEntries []Entry
  filteredEntries, err = GetFilteredEntries(entries, project, task, sinceTime, untilTime)
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }

  if listOnlyProjectsAndTasks || listOnlyTasks {
    printProjects(filteredEntries)
    return nil
  }
  return filteredEntries
}

func printProjects(entries []Entry) {

  projectsAndTasks, _ := listProjectsAndTasks(entries)
  for project := range projectsAndTasks {
    if listOnlyProjectsAndTasks && !listOnlyTasks {
      fmt.Printf("%s %s\n", CharMore, project)
    }

    for task := range projectsAndTasks[project] {
      if listOnlyProjectsAndTasks && !listOnlyTasks {
        fmt.Printf("%*s└── ", 1, " ")
      }

      if appendProjectIDToTask {
        fmt.Printf("%s [%s]\n", task, project)
      } else {
        fmt.Printf("%s\n", task)
      }
    }
  }
}

func listProjectsAndTasks(entries []Entry) (map[string]map[string]bool, []string) {
  var projectsAndTasks = make(map[string]map[string]bool)
  var allTasks []string

  for _, filteredEntry := range entries {
    taskMap, ok := projectsAndTasks[filteredEntry.Project]

    if !ok {
      taskMap = make(map[string]bool)
      projectsAndTasks[filteredEntry.Project] = taskMap
    }

    taskMap[filteredEntry.Task] = true
    projectsAndTasks[filteredEntry.Project] = taskMap
    allTasks = append(allTasks, filteredEntry.Task)
  }

  return projectsAndTasks, allTasks
}

func trackTask() {
  user := GetCurrentUser()

  runningEntryId, err := database.GetRunningEntryId(user)
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }

  if runningEntryId != "" {
    fmt.Printf("%s a task is already running\n", CharTrack)
    os.Exit(1)
  }

  if project == "" && viper.GetString("project.default") != "" {
    project = viper.GetString("project.default")
  }

  if project == "" && viper.GetBool("project.mandatory") {
    fmt.Println("project is mandatory but missing")
    os.Exit(1)
  }

  if task == "" && viper.GetBool("task.mandatory") {
    fmt.Println("task is mandatory but missing")
    os.Exit(1)
  }

  newEntry, err := NewEntry("", begin, finish, project, task, user)
  if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
  }

  if notes != "" {
    newEntry.Notes = notes
  }

  isRunning := newEntry.Finish.IsZero()

  _, err = database.AddEntry(user, newEntry, isRunning)
  if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
  }

  fmt.Print(newEntry.GetOutputForTrack(isRunning, false))
}

func finishTask(mode int) {

  user := GetCurrentUser()

  runningEntryId, err := database.GetRunningEntryId(user)
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }

  if runningEntryId == "" {
    fmt.Printf("%s not running\n", CharFinish)
    os.Exit(1)
  }

  runningEntry, err := database.GetEntry(user, runningEntryId)
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }

  tmpEntry, err := NewEntry(runningEntry.ID, begin, finish, project, task, user)
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }

  if begin != "" {
    runningEntry.Begin = tmpEntry.Begin
  }

  if finish != "" {
    runningEntry.Finish = tmpEntry.Finish
  } else {
    runningEntry.Finish = time.Now()
  }

  if mode == FinishWithMetadata {
    finishTaskMetadata(user, &runningEntry, &tmpEntry)
  }

  if !runningEntry.IsFinishedAfterBegan() {
    fmt.Printf("%s %+v\n", CharError, "beginning time of tracking cannot be after finish time")
    os.Exit(1)
  }

  _, err = database.FinishEntry(user, runningEntry)
  if err != nil {
    fmt.Printf("%s %+v\n", CharError, err)
    os.Exit(1)
  }

  fmt.Print(runningEntry.GetOutputForFinish())
}

func finishTaskMetadata(user string, runningEntry *Entry, tmpEntry *Entry) {

  if project != "" {
    runningEntry.Project = tmpEntry.Project
  }

  if task != "" {
    runningEntry.Task = tmpEntry.Task
  }

  if notes != "" {
    runningEntry.Notes = fmt.Sprintf("%s\n%s", runningEntry.Notes, notes)
  }

  if runningEntry.Task != "" {
    task, err := database.GetTask(user, runningEntry.Task)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    taskGit(&task, runningEntry)
  }

}

func taskGit(task *Task, runningEntry *Entry) {
  if task.GitRepository != "" && task.GitRepository != "-" {
    stdout, stderr, err := GetGitLog(task.GitRepository, runningEntry.Begin, runningEntry.Finish)
    if err != nil {
      fmt.Printf("%s %+v\n", CharError, err)
      os.Exit(1)
    }

    if stderr == "" {
      runningEntry.Notes = fmt.Sprintf("%s\n%s", runningEntry.Notes, stdout)
    } else {
      fmt.Printf("%s notes were not imported: %+v\n", CharError, stderr)
    }
  }
}

