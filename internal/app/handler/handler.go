package handler

import (
	"rip/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	// Routes
	api.GET("/routes", h.GetAllRoutes)
	api.GET("/routes/:route_id", h.GetRouteByID)
	api.POST("/routes", h.CreateRoute)
	api.PUT("/routes/:route_id", h.UpdateRoute)
	api.DELETE("/routes/:route_id", h.DeleteRoute)
	api.POST("/routes/:route_id/image", h.UploadRouteImage)
	api.POST("/draft/addroute/:route_id", h.AddToDraft)

	// SpeedRequests
	api.GET("/speedrequests/draft", h.GetDraftInfo)
	api.GET("/speedrequests", h.GetAllSpeedRequests)
	api.GET("/speedrequests/:speed_request_id", h.GetSpeedRequestByID)
	api.PUT("/speedrequests/:speed_request_id", h.UpdateSpeedRequest)
	api.PUT("/speedrequests/:speed_request_id/submit", h.SubmitSpeedRequest)
	api.PUT("/speedrequests/:speed_request_id/complete", h.CompleteSpeedRequest)
	api.DELETE("/speedrequests/:speed_request_id", h.DeleteSpeedRequest)

	// RouteSpeedRequest
	api.DELETE("/routespeedrequest", h.RemoveRouteSpeedRequest)
	api.PUT("/routespeedrequest", h.UpdateRouteSpeedRequest)

	// Auth
	api.POST("/auth/signup", h.SignUp)
	api.POST("/auth/signin", h.SignIn)
	api.POST("/auth/signout", h.SignOut)
	api.GET("/users/me", h.GetCurrentUser)
	api.PUT("/users/me", h.UpdateUser)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.Static("/img", "./resources/img")

}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
