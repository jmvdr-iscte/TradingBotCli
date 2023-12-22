package initialize

import (
	"os"
)

type AlpacaConfig struct {
	ID     string
	Secret string
	Url    string
}

func LoadAlpaca() *AlpacaConfig {
	cfg := &AlpacaConfig{
		ID:     "",
		Secret: "",
		Url:    "",
	}

	if id, exists := os.LookupEnv("APCA_API_KEY_ID"); exists {
		cfg.ID = id
	}

	if secret, exists := os.LookupEnv("APCA_API_SECRET_KEY"); exists {
		cfg.Secret = secret
	}

	if url, exists := os.LookupEnv("APCA_API_BASE_URL"); exists {
		cfg.Url = url
	}

	return cfg
}
