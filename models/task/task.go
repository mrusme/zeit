package task

import (
	"errors"
	"strings"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/out"
)

var ErrSIDNotFound error = errors.New("SID not found")

type Task struct {
	key         string `json:"-"`
	OwnerKey    string `json:"owner_key"`
	SID         string `json:"sid"`
	DisplayName string `json:"display_name"`
	Color       string `json:"color"`
}

func New(ownerKey string, sid string) (*Task, error) {
	tk := new(Task)
	tk.key = database.NewKey(tk)
	tk.OwnerKey = ownerKey
	tk.SID = sid
	tk.DisplayName = strings.ToTitle(sid)
	tk.Color = out.RandomVsibleHexColor()
	return tk, nil
}

func (tk *Task) SetKey(k string) {
	tk.key = k
	return
}

func (tk *Task) GetKey() string {
	return tk.key
}

func List(db *database.Database) (map[string]*Task, error) {
	var err error

	var rows map[string]*Task = make(map[string]*Task)
	if err = database.GetPrefixedRowsAsStruct(
		db,
		database.PrefixForModel(&Task{}),
		rows,
	); err != nil {
		return nil, err
	}

	return rows, nil
}

func Get(db *database.Database, sid string) (*Task, error) {
	var err error

	var rows map[string]*Task = make(map[string]*Task)
	if rows, err = List(db); err != nil {
		return nil, err
	}

	for _, tk := range rows {
		if tk.SID == sid {
			return tk, nil
		}
	}

	return nil, ErrSIDNotFound
}

func Set(db *database.Database, tk *Task) error {
	if err := db.UpsertRowAsStruct(tk); err != nil {
		return err
	}

	return nil
}

func InsertIfNone(db *database.Database, ownerKey string, sid string) (*Task, error) {
	var tk *Task
	var err error

	tk, err = Get(db, sid)
	if err != nil && err != ErrSIDNotFound {
		return nil, err
	} else if err != nil && err == ErrSIDNotFound {
		tk, _ = New(ownerKey, sid)
		if err = Set(db, tk); err != nil {
			return nil, err
		}
	}

	return tk, nil
}
