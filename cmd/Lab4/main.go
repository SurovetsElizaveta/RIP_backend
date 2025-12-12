package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"rip/internal/app/config"
	"rip/internal/app/dsn"
	"rip/internal/app/handler"
	"rip/internal/app/jwt"
	"rip/internal/app/redis"
	"rip/internal/app/repository"
	"rip/internal/app/service"
	"rip/internal/pkg/minio"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "rip/docs"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigins := []string{
			"http://localhost:3000",
			"https://localhost:3000",
			"https://surovetselizaveta.github.io",
			"http://10.165.215.65:3000",
			"https://10.165.215.65:3000",
		}

		origin := c.Request.Header.Get("Origin")

		for _, allowed := range allowedOrigins {
			if allowed == origin {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

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
	router.Use(CORSMiddleware())

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

	asyncService := service.NewAsyncService(conf.AsyncService)
	hand := handler.NewHandler(rep, jwtManager, redisClient, asyncService, conf)

	hand.RegisterHandler(router)
	hand.RegisterStatic(router)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort),
		Handler: router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	logrus.Infof("Starting HTTPS server on https://%s:%d", conf.ServiceHost, conf.ServicePort)

	err = server.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		logrus.Fatalf("Failed to start HTTPS server: %v", err)
	}
}
