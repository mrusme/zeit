package database

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
)

func (db *Database) UpsertRowAsBytes(key string, val []byte) error {
	err := db.engine.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), val)
		return err
	})
	return err
}

func (db *Database) UpsertRowAsString(key string, val string) error {
	return db.UpsertRowAsBytes(key, []byte(val))
}

func (db *Database) UpsertRowAsStruct(v Model) error {
	val, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return db.UpsertRowAsBytes(v.GetKey(), val)
}
