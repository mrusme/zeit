package database

import (
	"github.com/dgraph-io/badger/v4"
)

func (db *Database) DestroyRow(key string) error {
	err := db.engine.View(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	return err
}
