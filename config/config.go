package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string   `yaml:"port"`
	JWTSecret      string   `yaml:"jwtSecret"`
	DBPath         string   `yaml:"dbPath"`
	ApiGemini      string   `yaml:"apiGeminiKey"`
	Env            string   `yaml:"env"`            // dev, release
	GinMode        string   `yaml:"ginMode"`        // debug, release
	TrustedProxies []string `yaml:"trustedProxies"` // список доверенных IP/сетей
	LogLevel       string   `yaml:"logLevel"`       // debug, info, warn, error
	LogFile        string   `yaml:"logFile"`        // путь к файлу логов (если пусто - логи в stdout)
}

func LoadConfig() *Config {
	// Загружаем .env файл (если существует, ошибка игнорируется)
	_ = godotenv.Load(".env")

	cfg := &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		DBPath:    getEnv("DB_PATH", "data.db"),
		ApiGemini: getEnv("GEMINI_API_KEY", ""),
		Env:       getEnv("ENV", "dev"), // dev или release
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFile:   getEnv("LOG_FILE", ""),
	}

	// Определяем Gin mode в зависимости от ENV
	if cfg.Env == "release" {
		cfg.GinMode = "release"
	} else {
		cfg.GinMode = "debug"
	}

	// Парсим trusted proxies из env
	proxyStr := getEnv("TRUSTED_PROXIES", "")
	if proxyStr != "" {
		cfg.TrustedProxies = strings.Split(proxyStr, ",")
		for i := range cfg.TrustedProxies {
			cfg.TrustedProxies[i] = strings.TrimSpace(cfg.TrustedProxies[i])
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
