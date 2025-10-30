package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ServerPort      string
	DBpediaEndpoint string
	ReleaseMode     bool
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	config := &Config{
		ServerPort:      getEnv("SERVER_PORT", ":8080"),
		DBpediaEndpoint: getEnv("DBPEDIA_ENDPOINT", "https://dbpedia.org/sparql"),
		ReleaseMode:     getEnv("GIN_MODE", "release") == "release",
	}

	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
