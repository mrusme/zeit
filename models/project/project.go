package project

import (
	"strings"

	"github.com/mrusme/zeit/database"
	"github.com/mrusme/zeit/errs"
	"github.com/mrusme/zeit/helpers/out"
)

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

func List(db *database.Database) (map[string]*Project, error) {
	var err error

	var rows map[string]*Project = make(map[string]*Project)
	if err = database.GetPrefixedRowsAsStruct(
		db,
		database.PrefixForModel(&Project{}),
		rows,
	); err != nil {
		return nil, err
	}

	return rows, nil
}

func Get(db *database.Database, key string) (*Project, error) {
	var err error

	pj := new(Project)
	err = db.GetRowAsStruct(key, pj)
	if err != nil && db.IsErrKeyNotFound(err) == false {
		// We encountered an error which is not KeyNotFound
		return nil, err
	} else if err != nil && db.IsErrKeyNotFound(err) == true {
		return nil, errs.ErrKeyNotFound
	}

	return pj, nil
}

func GetBySID(db *database.Database, sid string) (*Project, error) {
	var err error

	var rows map[string]*Project = make(map[string]*Project)
	if rows, err = List(db); err != nil {
		return nil, err
	}

	for _, pj := range rows {
		if pj.SID == sid {
			return pj, nil
		}
	}

	return nil, errs.ErrSIDNotFound
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

	pj, err = GetBySID(db, sid)
	if err != nil && err != errs.ErrSIDNotFound {
		return nil, err
	} else if err != nil && err == errs.ErrSIDNotFound {
		pj, _ = New(ownerKey, sid)
		if err = Set(db, pj); err != nil {
			return nil, err
		}
	}

	return pj, nil
}
