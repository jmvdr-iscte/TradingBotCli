package initialize

import (
	"os"
)

type RedisConfig struct {
	Address  string
	Password string
}

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
