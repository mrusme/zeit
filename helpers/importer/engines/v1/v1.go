package v1

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/errs"
	"github.com/mrusme/zeit/models/activeblock"
	"github.com/mrusme/zeit/models/block"
	"github.com/mrusme/zeit/models/config"
	"github.com/mrusme/zeit/models/project"
	"github.com/mrusme/zeit/models/task"
)

type V1 struct {
	fd *os.File
}

type Entries map[string]json.RawMessage

func New(fd *os.File) (*V1, error) {
	engine := new(V1)
	engine.fd = fd

	return engine, nil
}

func (engine *V1) Import(
	cb func(database.Model, error, ...any) error,
	v ...any,
) error {
	var err error

	decoder := json.NewDecoder(engine.fd)

	for {
		var entries Entries
		if err = decoder.Decode(&entries); err != nil {
			if err.Error() == "EOF" {
				break
			}

			if err = cb(nil, err, v...); err != nil {
				return err
			}
		}

		for key := range entries {
			entry := entries[key]

			modelName, _, found := strings.Cut(key, ":")
			if found == false {
				modelName = key
			}

			switch modelName {
			case "activeblock":
				var model activeblock.ActiveBlock
				if err = json.Unmarshal(entry, &model); err != nil {
					err = cb(nil, errs.ErrDataConversion, v...)
				}
				model.SetKey(key)
				err = cb(&model, nil, v...)
			case "block":
				var model block.Block
				if err = json.Unmarshal(entry, &model); err != nil {
					err = cb(nil, errs.ErrDataConversion, v...)
				}
				model.SetKey(key)
				err = cb(&model, nil, v...)
			case "config":
				var model config.Config
				if err = json.Unmarshal(entry, &model); err != nil {
					err = cb(nil, errs.ErrDataConversion, v...)
				}
				model.SetKey(key)
				err = cb(&model, nil, v...)
			case "project":
				var model project.Project
				if err = json.Unmarshal(entry, &model); err != nil {
					err = cb(nil, errs.ErrDataConversion, v...)
				}
				model.SetKey(key)
				err = cb(&model, nil, v...)
			case "task":
				var model task.Task
				if err = json.Unmarshal(entry, &model); err != nil {
					err = cb(nil, errs.ErrDataConversion, v...)
				}
				model.SetKey(key)
				err = cb(&model, nil, v...)
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}
