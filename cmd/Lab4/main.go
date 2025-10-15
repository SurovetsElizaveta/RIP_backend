package main

import (
	"rip/internal/app/config"
	"rip/internal/app/dsn"
	"rip/internal/app/handler"
	"rip/internal/app/jwt"
	"rip/internal/app/redis"
	"rip/internal/app/repository"
	"rip/internal/pkg"
	"rip/internal/pkg/minio"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()

	// Загружаем конфиг
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Инициализируем Minio
	minioConfig := config.NewMinioConfig()
	minioClient, err := minio.NewMinioClient(minioConfig)
	if err != nil {
		logrus.Fatalf("error initializing Minio client: %v", err)
	}

	// Инициализируем Redis
	redisClient, err := redis.New(conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing Redis client: %v", err)
	}

	// Инициализируем JWT менеджер
	jwtManager := jwt.NewManager(conf.JWT)

	postgresString := dsn.FromEnv()

	// Передаем зависимости в репозиторий
	rep, errRep := repository.New(postgresString, minioClient)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep, jwtManager, redisClient)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
