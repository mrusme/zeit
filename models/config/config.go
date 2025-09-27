package config

import (
	"github.com/google/uuid"
	"github.com/mrusme/zeit/database"
)

const KEY string = "config"

type Config struct {
	key     string
	UserKey string `json:"user_key"`
}

func New() (*Config, error) {
	var err error

	var userKey uuid.UUID
	if userKey, err = uuid.NewV7(); err != nil {
		return nil, err
	}

	cfg := new(Config)
	cfg.key = KEY
	cfg.UserKey = userKey.String()

	return cfg, err
}

func (cfg *Config) SetKey(k string) {
	cfg.key = k
}

func (cfg *Config) GetKey() string {
	return cfg.key
}

func Get(db *database.Database) (*Config, error) {
	var err error

	cfg, err := New()
	if err != nil {
		return nil, err
	}
	err = db.GetRowAsStruct(cfg.GetKey(), cfg)
	if err != nil && db.ErrIsKeyNotFound(err) == false {
		// We encountered an error which is not KeyNotFound
		return nil, err
	}

	// First time users won't have a Config, hence we will retrieve an error
	// that is of type KeyNotFound. In that case we would return a New()
	// Config, which just so happens to be in `cfg` anyway, hence we don't
	// need to handle that case.
	return cfg, nil
}

func Set(db *database.Database, cfg *Config) error {
	if err := db.UpsertRowAsStruct(cfg); err != nil {
		return err
	}

	return nil
}
