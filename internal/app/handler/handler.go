package handler

import (
	"rip/internal/app/jwt"
	"rip/internal/app/redis"
	"rip/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	JWTManager *jwt.Manager
	Redis      *redis.Client
}

func NewHandler(r *repository.Repository, jwtManager *jwt.Manager, redisClient *redis.Client) *Handler {
	return &Handler{
		Repository: r,
		JWTManager: jwtManager,
		Redis:      redisClient,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	// Public routes
	api.POST("/signup", h.SignUp)
	api.POST("/signin", h.SignIn)
	api.POST("/refresh", h.RefreshToken)

	api.GET("/routes", h.GetAllRoutes)
	api.GET("/routes/:route_id", h.GetRouteByID)

	api.GET("/speedrequests/draft", h.GetDraftInfo)

	// Protected routes
	auth := api.Group("/")
	auth.Use(h.AuthMiddleware())
	{
		auth.POST("/signout", h.SignOut)
		auth.GET("/users/me", h.GetCurrentUser)
		auth.PUT("/users/me", h.UpdateUser)

		auth.POST("/draft/addroute/:route_id", h.AddToDraft)

		auth.GET("/speedrequests", h.GetAllSpeedRequests)
		auth.GET("/speedrequests/:speed_request_id", h.GetSpeedRequestByID)
		auth.PUT("/speedrequests/:speed_request_id", h.UpdateSpeedRequest)
		auth.PUT("/speedrequests/:speed_request_id/submit", h.SubmitSpeedRequest)
		auth.DELETE("/speedrequests/:speed_request_id", h.DeleteSpeedRequest)

		auth.DELETE("/routespeedrequest", h.RemoveRouteSpeedRequest)
		auth.PUT("/routespeedrequest", h.UpdateRouteSpeedRequest)

		// Moderator
		moderator := auth.Group("/")
		moderator.Use(h.ModeratorMiddleware())
		{
			moderator.PUT("/speedrequests/:speed_request_id/complete", h.CompleteSpeedRequest)
			moderator.POST("/routes", h.CreateRoute)
			moderator.PUT("/routes/:route_id", h.UpdateRoute)
			moderator.DELETE("/routes/:route_id", h.DeleteRoute)
			moderator.POST("/routes/:route_id/image", h.UploadRouteImage)
		}
	}
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.Static("/img", "./resources/img")

}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
