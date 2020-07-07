package config

import (
	"os"
	"strconv"
	"strings"
)

// Server configs
type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// Database configs
type Database struct {
	Connection string `json:"connection"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Db         string `json:"database"`
	User       string `json:"user"`
	Pass       string `json:"pass"`
}

// LocalConfig ...
type LocalConfig struct {
	APPKey    []byte `json:"app_key"`
	DebugMode bool   `json:"debug_mode"`
	Server    `json:"server"`
	Database  `json:"database"`
}

// New - returns a new Local struct
func New() *LocalConfig {
	return &LocalConfig{
		APPKey:    []byte(getEnv("APP_KEY", "")),
		DebugMode: getEnvAsBool("DEBUG_MODE", true),
		Server: Server{
			Host: getEnv("SRV_HOST", ""),
			Port: getEnv("SRV_PORT", ""),
		},
		Database: Database{
			Connection: getEnv("DB_CONNECTION", "postgres"),
			Host:       getEnv("DB_HOST", "localhost"),
			Port:       getEnvAsInt("DB_PORT", 5432),
			Db:         getEnv("DB_DATABASE", ""),
			User:       getEnv("DB_USERNAME", ""),
			Pass:       getEnv("DB_PASSWORD", ""),
		},
	}
}

// Getting a env-variable
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Getting a env-variable of type int
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Getting a env-variable of type bool
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Getting a env-variable of type slice
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
