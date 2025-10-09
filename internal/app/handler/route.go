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

func (h *Handler) CreateRoute(ctx *gin.Context) {
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

func (h *Handler) UpdateRoute(ctx *gin.Context) {
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

func (h *Handler) DeleteRoute(ctx *gin.Context) {
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

func (h *Handler) AddToDraft(ctx *gin.Context) {
	routeIDStr := ctx.Param("route_id")
	routeID, err := strconv.Atoi(routeIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID маршрута"))
		return
	}

	currentUserID := uint(1)

	speedRequestID, err := h.Repository.AddRouteToDraft(uint(routeID), currentUserID)
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

func (h *Handler) UploadRouteImage(ctx *gin.Context) {
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
