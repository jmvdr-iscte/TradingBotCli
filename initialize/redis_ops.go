// Package initialize serves to initialize the configs.
package initialize

import (
	"os"
)

// RedisConfig is the initial Alpaca api config.
type RedisConfig struct {
	Address  string
	Password string
}

// LoadRedisConfigs loads the redis configs with the values from .env.
func LoadRedisConfigs() *RedisConfig {
	cfg := &RedisConfig{
		Address:  "",
		Password: "",
	}

	if redis_address, exists := os.LookupEnv("REDIS_ADDR"); exists {
		cfg.Address = redis_address
	}

	if redis_password, exists := os.LookupEnv("DB_PASSWORD"); exists {
		cfg.Password = redis_password
	}
	return cfg
}
