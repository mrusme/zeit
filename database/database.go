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

func New(logger badger.Logger, dbpath string, readOnly bool) (*Database, error) {
	var err error

	bopt := badger.DefaultOptions(dbpath).
		WithLogger(logger).
		WithReadOnly(readOnly)
		// WithChecksumVerificationMode(options.OnTableAndBlockRead)

	// if encrypted {
	// 	bopt = bopt.WithEncryptionKey("").WithIndexCacheSize(100 << 20)
	// }
	if dbpath == "" {
		bopt = bopt.WithInMemory(true)
	}

	db := new(Database)
	db.logger = logger
	if db.engine, err = badger.Open(bopt); err == badger.ErrWindowsNotSupported || err == badger.ErrPlan9NotSupported {
		bopt = bopt.WithReadOnly(false)
		db.engine, err = badger.Open(bopt)
	}
	if err != nil {
		return db, err
	}

	return db, nil
}

func (db *Database) Close() {
	for {
		err := db.engine.RunValueLogGC(0.7)
		if err == nil {
			continue
		}

		if errors.Is(err, badger.ErrNoRewrite) == false &&
			errors.Is(err, badger.ErrGCInMemoryMode) == false &&
			errors.Is(err, badger.ErrGCInReadOnlyMode) == false {
			db.logger.Warningf("Value log GC stopped: %s", err)
		}

		break
	}

	if err := db.engine.Close(); err != nil {
		db.logger.Errorf("Error closing database: %s", err)
	}
}

func (db *Database) IsErrKeyNotFound(err error) bool {
	return errors.Is(err, badger.ErrKeyNotFound)
}
