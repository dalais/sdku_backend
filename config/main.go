package config

import (
	"os"
	"strconv"
	"strings"
)

// LocalConfig ...
type LocalConfig struct {
	APPKey       string
	DebugMode    bool
	DbConnection string
	DbHost       string
	DbPort       int
	DbDatabase   string
	DbUsername   string
	DbPassword   string
}

// New returns a new Local struct
func New() *LocalConfig {
	return &LocalConfig{
		APPKey:       getEnv("APP_KEY", ""),
		DebugMode:    getEnvAsBool("DEBUG_MODE", true),
		DbConnection: getEnv("DB_CONNECTION", "postgres"),
		DbHost:       getEnv("DB_HOST", "localhost"),
		DbPort:       getEnvAsInt("DB_PORT", 5432),
		DbDatabase:   getEnv("DB_DATABASE", ""),
		DbUsername:   getEnv("DB_USERNAME", ""),
		DbPassword:   getEnv("DB_PASSWORD", ""),
	}
}

// Получение env-переменной
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Получение env-переменной типа int
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Получение env-переменной типа bool
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Получение env-переменной типа slice
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
