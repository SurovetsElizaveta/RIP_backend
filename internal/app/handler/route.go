package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"rip/internal/app/ds"
	"rip/internal/app/dto"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetAllRoutes godoc
// @Summary Get all routes
// @Description Get all routes list with optional filters
// @Tags routes
// @Produce json
// @Param min_distance query int false "Minimum distance filter"
// @Param max_distance query int false "Maximum distance filter"
// @Success 200 {array} ds.Route "routes"
// @Failture 500
// @Router /routes [get]
func (h *Handler) GetAllRoutes(ctx *gin.Context) {
	var routes []ds.Route
	var err error

	minDistanceStr := ctx.Query("min_distance")
	maxDistanceStr := ctx.Query("max_distance")
	if minDistanceStr == "" && maxDistanceStr == "" {
		routes, err = h.Repository.GetAllRoutes()
		if err != nil {
			logrus.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		var minDistance, maxDistance int
		var err error

		if minDistanceStr != "" {
			minDistance, err = strconv.Atoi(minDistanceStr)
			if err != nil {
				logrus.Error("Ошибка преобразования min_distance:", err)
				minDistance = 0
			}
		}

		if maxDistanceStr != "" {
			maxDistance, err = strconv.Atoi(maxDistanceStr)
			if err != nil {
				logrus.Error("Ошибка преобразования max_distance:", err)
				maxDistance = 10000
			}
		}

		if minDistanceStr == "" {
			minDistance = 0
		}
		if maxDistanceStr == "" {
			maxDistance = 20000
		}

		routes, err = h.Repository.GetRoutesByDistance(minDistance, maxDistance)
		if err != nil {
			logrus.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, routes)
}

// GetRouteByID godoc
// @Summary Get route by ID
// @Description Get route information by ID
// @Tags routes
// @Produce json
// @Param route_id path int true "Route ID"
// @Success 200 {object} ds.Route "route"
// @Failture 500
// @Router /routes/{route_id} [get]
func (h *Handler) GetRouteByID(ctx *gin.Context) {
	strId := ctx.Param("route_id")
	route_id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	route, err := h.Repository.GetRouteByID(uint(route_id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, route)
}

// CreateRoute godoc
// @Summary Create new route
// @Description Create new route. Only for moderator
// @Tags routes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ds.Route "route"
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /routes [post]
func (h *Handler) CreateRoute(ctx *gin.Context) {
	isModerator, exists := ctx.Get("is_moderator")
	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

	if !isModerator.(bool) {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("only moderators can create routes"))
		return
	}

	var request dto.CreateRoute

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	route := &ds.Route{
		Title:       request.Title,
		Distance:    request.Distance,
		Description: request.Description,
		Delay:       request.Delay,
		Status:      "действует",
	}

	if err := h.Repository.CreateRoute(route); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, route)
}

// UpdateRoute godoc
// @Summary Update existing route
// @Description Update existing route. Only for moderator
// @Tags routes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ds.Route "route"
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /routes/{route_id} [put]
func (h *Handler) UpdateRoute(ctx *gin.Context) {
	isModerator, exists := ctx.Get("is_moderator")
	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

	if !isModerator.(bool) {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("only moderators can create routes"))
		return
	}

	routeID, err := strconv.Atoi(ctx.Param("route_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("неверный ID маршрута"))
		return
	}

	var request dto.UpdateRoute
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	existingRoute, err := h.Repository.GetRouteByID(uint(routeID))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if request.Title != nil {
		existingRoute.Title = *request.Title
	}
	if request.Distance != nil {
		existingRoute.Distance = *request.Distance
	}
	if request.Description != nil {
		existingRoute.Description = *request.Description
	}
	if request.Delay != nil {
		existingRoute.Delay = *request.Delay
	}
	if request.Status != nil {
		existingRoute.Status = *request.Status
	}

	if err := h.Repository.UpdateRoute(existingRoute); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, existingRoute)
}

// DeleteRoute godoc
// @Summary Delete existing route
// @Description Delete existing route. Only for moderator
// @Tags routes
// @Produce json
// @Security BearerAuth
// @Param route_id path int true "Route ID"
// @Success 200 {object} string "Маршрут успешно удален"
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /routes/{route_id} [delete]
func (h *Handler) DeleteRoute(ctx *gin.Context) {
	isModerator, exists := ctx.Get("is_moderator")
	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

	if !isModerator.(bool) {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("only moderators can create routes"))
		return
	}

	routeID, err := strconv.Atoi(ctx.Param("route_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("неверный ID маршрута"))
		return
	}

	if err := h.Repository.DeleteRoute(uint(routeID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Маршрут успешно удален",
	})
}

// AddToDraft godoc
// @Summary Add route to draft
// @Description Add route to draft. Create new darft if darft not existing. For authentificated users (client, moderator)
// @Tags routes
// @Produce json
// @Security BearerAuth
// @Param route_id path int true "Route ID"
// @Success 200 {object} object
// @Failture 404
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /draft/addroute/{route_id} [post]
func (h *Handler) AddToDraft(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")

	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

	routeIDStr := ctx.Param("route_id")
	routeID, err := strconv.Atoi(routeIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID маршрута"))
		return
	}

	speedRequestID, err := h.Repository.AddRouteToDraft(uint(routeID), currentUserID.(uint))
	if err != nil {
		if err.Error() == "маршрут не найден" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if err.Error() == "маршрут уже добавлен в заявку" {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else {
			logrus.Errorf("Ошибка добавления в черновик: %v", err)
			h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка добавления в заявку"))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":          "Маршрут успешно добавлен в заявку",
		"speed_request_id": speedRequestID,
	})
}

// UploadRouteImage godoc
// @Summary Upload route image
// @Description Upload route image. Add new image if route doesn`t have one or change image. Only for moderator
// @Tags routes
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param route_id path int true "Route ID"
// @Success 200 {object} string "Изображение успешно загружено и обновлено"
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /routes/{route_id}/image [post] [get]
func (h *Handler) UploadRouteImage(ctx *gin.Context) {
	isModerator, exists := ctx.Get("is_moderator")
	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

	if !isModerator.(bool) {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("only moderators can create routes"))
		return
	}
	routeIDStr := ctx.Param("route_id")
	routeID, err := strconv.ParseUint(routeIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID маршрута"))
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("файл изображения обязателен"))
		return
	}

	if err := h.validateImageFile(file); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UploadRouteImage(uint(routeID), file); err != nil {
		logrus.Errorf("Ошибка загрузки изображения: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка загрузки изображения: %v", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Изображение успешно загружено и обновлено",
	})
}

func (h *Handler) validateImageFile(file *multipart.FileHeader) error {
	const maxSize = 5 << 20
	if file.Size > maxSize {
		return fmt.Errorf("размер файла не должен превышать 5MB")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
		".svg":  true,
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("разрешены только файлы с расширениями: jpg, jpeg, png, gif, webp, bmp, svg")
	}

	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/bmp":  true,
		"image/svg":  true,
	}

	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	case ".bmp":
		mimeType = "image/bmp"
	case ".svg":
		mimeType = "image/svg"
	default:
		return fmt.Errorf("неподдерживаемый формат изображения")
	}

	if !allowedMimeTypes[mimeType] {
		return fmt.Errorf("неподдерживаемый MIME type изображения")
	}

	return nil
}
