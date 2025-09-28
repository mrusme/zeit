package task

import (
	"strings"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/errs"
	"github.com/mrusme/zeit/helpers/out"
	"github.com/mrusme/zeit/helpers/val"
)

type Task struct {
	key         string `json:"-"`
	OwnerKey    string `json:"owner_key"`
	SID         string `json:"sid" validate:"required,sid,max=32"`
	ProjectSID  string `json:"project_sid" validate:"required,sid,max=32"`
	DisplayName string `json:"display_name" validate:"max=32"`
	Color       string `json:"color" validate:"hexcolor"`
}

func New(ownerKey string, projectSID string, sid string) (*Task, error) {
	tk := new(Task)
	tk.key = database.NewKey(tk)
	tk.OwnerKey = ownerKey
	tk.SID = sid
	tk.ProjectSID = projectSID
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

func ListForProjectSID(
	db *database.Database,
	projectSID string,
) (map[string]*Task, error) {
	var lst map[string]*Task = make(map[string]*Task)

	rows, err := List(db)
	if err != nil {
		return nil, err
	}

	for key := range rows {
		if rows[key].ProjectSID == projectSID {
			lst[key] = rows[key]
		}
	}

	return lst, nil
}

func Get(db *database.Database, key string) (*Task, error) {
	var err error

	tk := new(Task)
	err = db.GetRowAsStruct(key, tk)
	if err != nil && db.IsErrKeyNotFound(err) == false {
		// We encountered an error which is not KeyNotFound
		return nil, err
	} else if err != nil && db.IsErrKeyNotFound(err) == true {
		return nil, errs.ErrKeyNotFound
	}

	return tk, nil
}

func GetBySID(db *database.Database, projectSID string, sid string) (*Task, error) {
	var err error

	var rows map[string]*Task = make(map[string]*Task)
	if rows, err = List(db); err != nil {
		return nil, err
	}

	for _, tk := range rows {
		if tk.ProjectSID == projectSID && tk.SID == sid {
			return tk, nil
		}
	}

	return nil, errs.ErrSIDNotFound
}

func Set(db *database.Database, tk *Task) error {
	var err error

	if err = val.Validate(*tk); err != nil {
		return err
	}

	if err = db.UpsertRowAsStruct(tk); err != nil {
		return err
	}

	return nil
}

func InsertIfNone(db *database.Database, ownerKey string, projectSID string, sid string) (*Task, error) {
	var tk *Task
	var err error

	tk, err = GetBySID(db, projectSID, sid)
	if err != nil && err != errs.ErrSIDNotFound {
		return nil, err
	} else if err != nil && err == errs.ErrSIDNotFound {
		tk, _ = New(ownerKey, projectSID, sid)
		if err = Set(db, tk); err != nil {
			return nil, err
		}
	}

	return tk, nil
}
