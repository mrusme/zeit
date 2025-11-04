package v0

import (
	"encoding/json"
	"os"
	"time"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/val"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/models/project"
	"github.com/mrusme/zeit/models/task"
)

type V0 struct {
	fd *os.File
}

type Entry struct {
	Begin   time.Time `json:"begin"`
	Finish  time.Time `json:"finish"`
	Project string    `json:"project"`
	Task    string    `json:"task"`
	User    string    `json:"user"`
}

func New(fd *os.File) (*V0, error) {
	engine := new(V0)
	engine.fd = fd

	return engine, nil
}

func (engine *V0) Import(
	cb func(database.Model, error, ...any) error,
	v ...any,
) error {
	var b *block.Block
	var err error

	decoder := json.NewDecoder(engine.fd)

	var projectMap map[string]string = make(map[string]string)
	var taskMap map[string]string = make(map[string]string)

	for {
		var entries []Entry
		if err = decoder.Decode(&entries); err != nil {
			if err.Error() == "EOF" {
				break
			}

			if err = cb(nil, err, v...); err != nil {
				return err
			}
		}

		for i := range entries {
			var ok bool

			entry := entries[i]

			if b, err = block.New(""); err != nil {
				if err = cb(nil, err, v...); err != nil {
					return err
				}
			}

			var projectSID string
			if projectSID, ok = projectMap[entry.Project]; !ok {
				var pj *project.Project

				if entry.Project == "" {
					entry.Project = "undefined"
				}
				projectSID = val.ConvertTextToSID(entry.Project)
				if pj, err = project.New("", projectSID); err != nil {
					if err = cb(nil, err, v...); err != nil {
						return err
					}
				} else {
					pj.DisplayName = val.FitDisplayName(entry.Project)

					if err = cb(pj, nil, v...); err != nil {
						return err
					}

					b.ProjectSID = pj.SID
					projectMap[entry.Project] = pj.SID
				}
			} else {
				b.ProjectSID = projectSID
			}

			var taskSID string
			if taskSID, ok = taskMap[entry.Task]; !ok {
				var tk *task.Task

				if entry.Task == "" {
					entry.Task = "undefined"
				}
				taskSID = val.ConvertTextToSID(entry.Task)
				if tk, err = task.New("", projectSID, taskSID); err != nil {
					if err = cb(nil, err, v...); err != nil {
						return err
					}
				} else {
					tk.ProjectSID = projectSID
					tk.DisplayName = val.FitDisplayName(entry.Task)

					if err = cb(tk, nil, v...); err != nil {
						return err
					}

					b.TaskSID = tk.SID
					taskMap[entry.Task] = tk.SID
				}
			} else {
				b.TaskSID = taskSID
			}

			b.TimestampStart = entry.Begin
			b.TimestampEnd = entry.Finish

			if err = cb(b, nil, v...); err != nil {
				return err
			}
		}
	}

	return nil
}
