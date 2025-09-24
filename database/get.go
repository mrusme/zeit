package database

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"
)

func (db *Database) GetRowAsBytes(key string) ([]byte, error) {
	var ret []byte
	err := db.engine.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			ret = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})

	return ret, err
}

func (db *Database) GetRowAsString(key string) (string, error) {
	b, err := db.GetRowAsBytes(key)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (db *Database) GetRowAsStruct(key string, v Model) error {
	data, err := db.GetRowAsBytes(key)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, v); err != nil {
		return err
	}

	v.SetKey(key)
	return nil
}

func (db *Database) GetPrefixedRowsAsBytes(prefix string) (map[string][]byte, error) {
	var ret map[string][]byte = make(map[string][]byte)
	err := db.engine.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				ret[string(key)] = val
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return ret, err
}

func GetPrefixedRowsAsStruct[T Model](db *Database, prefix string, v map[string]T) error {
	errstr := ""

	rows, err := db.GetPrefixedRowsAsBytes(prefix)
	if err != nil {
		return err
	}

	for key := range rows {
		var t T
		if err := json.Unmarshal(rows[key], &t); err != nil {
			return err
		}
		t.SetKey(key)
		v[key] = t
	}

	if errstr != "" {
		return errors.New(errstr)
	}
	return nil
}
