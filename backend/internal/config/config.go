package config

import "os"

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	AppName  string
	Port     string
	LogLevel string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisHost string
	RedisPort string

	QwenAPIKey  string
	QwenBaseURL string
	QwenModel   string
	JWTSecret   string
}

// Load reads configuration from environment variables, applying sane
// defaults for local development where a value is not set.
func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "OpsPilot AI"),
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "opspilot"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		QwenAPIKey:  getEnv("QWEN_API_KEY", ""),
		QwenBaseURL: getEnv("QWEN_BASE_URL", "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"),
		QwenModel:   getEnv("QWEN_MODEL", "qwen-plus"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
