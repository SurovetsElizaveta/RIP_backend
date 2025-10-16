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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "rip/docs"
)

// @title RIP API
// @version 1.0
// @description Route Information Platform API
// @host localhost:8080
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	router := gin.Default()

	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"*"}, // Разрешить все origins для разработки
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	minioConfig := config.NewMinioConfig()
	minioClient, err := minio.NewMinioClient(minioConfig)
	if err != nil {
		logrus.Fatalf("error initializing Minio client: %v", err)
	}

	redisClient, err := redis.New(conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing Redis client: %v", err)
	}

	jwtManager := jwt.NewManager(conf.JWT)

	postgresString := dsn.FromEnv()

	rep, errRep := repository.New(postgresString, minioClient)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	hand := handler.NewHandler(rep, jwtManager, redisClient)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
