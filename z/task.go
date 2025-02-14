package z

import (
  "fmt"
  "os"
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


