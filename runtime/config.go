package runtime

import (
	"github.com/google/uuid"
)

var CONFIG_KEY string = "config"

type Config struct {
	key     string
	UserKey string `json:"user_key"`
}

func DefaultConfig() (*Config, error) {
	var err error

	var userKey uuid.UUID
	if userKey, err = uuid.NewV7(); err != nil {
		return nil, err
	}

	cfg := new(Config)
	cfg.UserKey = userKey.String()

	return cfg, err
}

func (cfg *Config) SetKey(k string) {
	cfg.key = k
}

func (cfg *Config) GetKey() string {
	return cfg.key
}
