package project

import (
	"errors"
	"strings"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/helpers/out"
)

var ErrSIDNotFound error = errors.New("SID not found")

type Project struct {
	key         string `json:"-"`
	OwnerKey    string `json:"owner_key"`
	SID         string `json:"sid"`
	DisplayName string `json:"display_name"`
	Color       string `json:"color"`
}

func New(ownerKey string, sid string) (*Project, error) {
	pj := new(Project)
	pj.key = database.NewKey(pj)
	pj.OwnerKey = ownerKey
	pj.SID = sid
	pj.DisplayName = strings.ToTitle(sid)
	pj.Color = out.RandomVsibleHexColor()
	return pj, nil
}

func (pj *Project) SetKey(k string) {
	pj.key = k
	return
}

func (pj *Project) GetKey() string {
	return pj.key
}

func Get(db *database.Database, sid string) (*Project, error) {
	var err error

	var rows map[string]*Project = make(map[string]*Project)
	if err = database.GetPrefixedRowsAsStruct(
		db,
		database.PrefixForModel(&Project{}),
		rows,
	); err != nil {
		return nil, err
	}

	for _, pj := range rows {
		if pj.SID == sid {
			return pj, nil
		}
	}

	return nil, ErrSIDNotFound
}

func Set(db *database.Database, pj *Project) error {
	if err := db.UpsertRowAsStruct(pj); err != nil {
		return err
	}

	return nil
}

func InsertIfNone(db *database.Database, ownerKey string, sid string) (*Project, error) {
	var pj *Project
	var err error

	pj, err = Get(db, sid)
	if err != nil && err != ErrSIDNotFound {
		return nil, err
	} else if err != nil && err == ErrSIDNotFound {
		pj, _ = New(ownerKey, sid)
		if err = Set(db, pj); err != nil {
			return nil, err
		}
	}

	return pj, nil
}
