package database

import (
	"errors"
	"runtime"

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

func New(logger badger.Logger, dbpath string, readOnly bool) (*Database, error) {
	var err error

	// badger does not support the readOnly option on all platforms
	const hasReadOnlySupport = runtime.GOOS != "windows" && runtime.GOOS != "plan9"

	bopt := badger.DefaultOptions(dbpath).
		WithLogger(logger).
		WithReadOnly(readOnly && hasReadOnlySupport)
		// WithChecksumVerificationMode(options.OnTableAndBlockRead)

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

func (db *Database) IsErrKeyNotFound(err error) bool {
	return errors.Is(err, badger.ErrKeyNotFound)
}
