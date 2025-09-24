package database

import (
	"errors"

	"github.com/dgraph-io/badger/v4"
)

type Database struct {
	logger badger.Logger
	engine *badger.DB
}

type Model interface {
	SetKey(string)
	GetKey() string
}

func New(logger badger.Logger, dbpath string) (*Database, error) {
	var err error

	bopt := badger.DefaultOptions(dbpath).
		WithLogger(logger)

	// if encrypted {
	// 	bopt = bopt.WithEncryptionKey("").WithIndexCacheSize(100 << 20)
	// }
	if dbpath == "" {
		bopt = bopt.WithInMemory(true)
	}

	db := new(Database)
	db.logger = logger
	if db.engine, err = badger.Open(bopt); err != nil {
		return db, err
	}

	return db, nil
}

func (db *Database) Close() {
	repeat := true

	for repeat {
		if err := db.engine.RunValueLogGC(0.7); err != nil {
			db.logger.Warningf("GC error: %s\n", err)
			repeat = false
		}
	}
	db.engine.Close()
}

func (db *Database) ErrIsKeyNotFound(err error) bool {
	return errors.Is(err, badger.ErrKeyNotFound)
}
