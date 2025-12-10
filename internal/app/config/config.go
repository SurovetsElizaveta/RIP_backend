package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost  string
	ServicePort  int
	JWT          JWTConfig
	Redis        RedisConfig
	AsyncService AsyncServiceConfig
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

type AsyncServiceConfig struct {
	URL   string
	Token string
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

	cfg.JWT.Secret = getEnv("JWT_SECRET", "your-super-secret-key")
	cfg.JWT.AccessTokenTTL = time.Hour * 1
	cfg.JWT.RefreshTokenTTL = time.Hour * 24 * 7

	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port, _ = strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB, _ = strconv.Atoi(getEnv("REDIS_DB", "0"))

	cfg.AsyncService.URL = getEnv("ASYNC_SERVICE_URL", "http://localhost:8000/api/calculate")
	cfg.AsyncService.Token = getEnv("ASYNC_SERVICE_TOKEN", "SECRET_ASYNC_TOKEN_DJANGO")

	logrus.Info("config parsed")

	return cfg, nil
}
