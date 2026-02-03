package config

import (
	"os"
)

type Config struct {
	Port  string
	DBUrl string
}

func Load() *Config {
	return &Config{
		Port:  getEnv("PORT", ":8080"),
		DBUrl: getEnv("DB_URL", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}
