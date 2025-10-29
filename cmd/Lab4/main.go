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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

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
	router.Use(CORSMiddleware())

	hand := handler.NewHandler(rep, jwtManager, redisClient)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
