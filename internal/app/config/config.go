package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	// JWT конфигурация из env
	cfg.JWT.Secret = getEnv("JWT_SECRET", "your-super-secret-key")
	cfg.JWT.AccessTokenTTL = time.Hour * 24      // 24 часа
	cfg.JWT.RefreshTokenTTL = time.Hour * 24 * 7 // 7 дней

	// Redis конфигурация из env
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port, _ = strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB, _ = strconv.Atoi(getEnv("REDIS_DB", "0"))

	log.Info("config parsed")

	return cfg, nil
}

// func getEnv(key, defaultValue string) string {
// 	if value := os.Getenv(key); value != "" {
// 		return value
// 	}
// 	return defaultValue
// }
