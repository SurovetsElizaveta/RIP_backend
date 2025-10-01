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
	maxDistanceStr := ctx.Query("max_distance") // получаем значение из поля поиска
	if minDistanceStr == "" && maxDistanceStr == "" {
		routes, err = h.Repository.GetAllRoutes()
		if err != nil {
			logrus.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Преобразуем параметры в числа
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
				maxDistance = 10000 // Большое значение по умолчанию
			}
		}

		// Если задан только один параметр, устанавливаем разумные значения по умолчанию
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

	// var speedDraft ds.SpeedRequest

	// speedDraft, err := h.Repository.GetDraftByUserID(uint(1))
	// if err != nil {
	// 	logrus.Error(err)
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// Преобразуем обратно в строки для отображения в форме
	// minDistanceValue := ctx.Query("min_distance")
	// maxDistanceValue := ctx.Query("max_distance")

	// var draftCount int64 = 0
	// if speedDraft != (ds.SpeedRequest{}) {
	// 	draftCount, err = h.Repository.GetSpeedRequestRoutesCount(speedDraft.SpeedRequestID)
	// 	if err != nil {
	// 		logrus.Error("Ошибка получения количества услуг в черновике:", err)
	// 		// Продолжаем выполнение с draftCount = 0
	// 	}
	// }

	// ctx.HTML(http.StatusOK, "routes.page.tmpl", gin.H{
	// 	"routes":       routes,
	// 	"minDistance":  minDistanceValue,
	// 	"maxDistance":  maxDistanceValue,
	// 	"draftCount":   draftCount,
	// 	"speedRequest": &speedDraft,
	// })

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
	// ctx.HTML(http.StatusOK, "route.page.tmpl", route)
}

func (h *Handler) CreateRoute(ctx *gin.Context) {
	var request dto.CreateRoute

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	route := ds.Route{
		Title:       request.Title,
		Distance:    request.Distance,
		Description: request.Description,
		Delay:       request.Delay,
		Status:      "действует",
		ImageURL:    "",
	}

	if err := h.Repository.CreateRoute(route); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// ctx.JSON(http.StatusCreated, gin.H{
	// 	"message": "Услуга успешно создана",
	// 	"id":      route.RouteID,
	// })

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

	// ctx.JSON(http.StatusOK, gin.H{
	// 	"message": "Услуга успешно обновлена",
	// })

	ctx.JSON(http.StatusOK, existingRoute)
}

func (h *Handler) DeleteRoute(ctx *gin.Context) {
	routeID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("неверный ID маршрута"))
		return
	}

	if err := h.Repository.DeleteRoute(uint(routeID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// ctx.JSON(http.StatusOK, gin.H{
	// 	"message": "Услуга успешно удалена",
	// })
}

func (h *Handler) UploadRouteImage(ctx *gin.Context) {
	// Получаем ID маршрута из URL параметра
	routeIDStr := ctx.Param("route_id") // обратите внимание на параметр :route_id
	routeID, err := strconv.ParseUint(routeIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID маршрута"))
		return
	}

	// Получаем файл из формы
	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("файл изображения обязателен"))
		return
	}

	// Валидация файла
	if err := h.validateImageFile(file); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Загружаем изображение через репозиторий
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
	// Проверяем размер файла (максимум 5MB)
	const maxSize = 5 << 20 // 5MB
	if file.Size > maxSize {
		return fmt.Errorf("размер файла не должен превышать 5MB")
	}

	// Проверяем расширение файла
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("разрешены только файлы с расширениями: jpg, jpeg, png, gif, webp, bmp")
	}

	// Проверяем MIME type (базовая проверка по расширению)
	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/bmp":  true,
	}

	// Определяем MIME type по расширению
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
	default:
		return fmt.Errorf("неподдерживаемый формат изображения")
	}

	if !allowedMimeTypes[mimeType] {
		return fmt.Errorf("неподдерживаемый MIME type изображения")
	}

	return nil
}

// #####################################################

// func (h *Handler) GetDraftByID(ctx *gin.Context) {
// 	speedRequestIDStr := ctx.Param("speed_request_id")
// 	speedRequestID, err := strconv.Atoi(speedRequestIDStr)
// 	if err != nil {
// 		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
// 		return
// 	}

// 	// Используем новый метод репозитория для поиска по ID
// 	speedRequest, err := h.Repository.GetSpeedRequestByID(speedRequestID)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "Заявка не найдена: " + err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	routes, err := h.Repository.GetRoutesBySpeedRequestID(speedRequest.SpeedRequestID)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	speedRequestRoutes, err := h.Repository.GetSpeedRequestRoutes(speedRequest.SpeedRequestID)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	speedRequestRoutesMap := make(map[int]ds.RouteSpeedRequest)
// 	for _, rr := range speedRequestRoutes {
// 		speedRequestRoutesMap[rr.RouteID] = rr
// 	}

// 	ctx.HTML(http.StatusOK, "draft.page.tmpl", gin.H{
// 		"speedRequest":          speedRequest,
// 		"routes":                routes,
// 		"speedRequestRoutesMap": speedRequestRoutesMap,
// 	})
// }

// func (h *Handler) AddToDraft(ctx *gin.Context) {
// 	routeIDStr := ctx.Param("route_id")
// 	routeID, err := strconv.Atoi(routeIDStr)
// 	if err != nil {
// 		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID маршрута"))
// 		return
// 	}

// 	// TODO: Получить userID из сессии (пока хардкод)
// 	userID := 1

// 	err = h.Repository.AddRouteToDraft(routeID, userID)
// 	if err != nil {
// 		// Перенаправляем с сообщением об ошибке
// 		ctx.Redirect(http.StatusSeeOther, "/routes?error="+url.QueryEscape(err.Error()))
// 		return
// 	}

// 	// Перенаправляем с сообщением об успехе
// 	ctx.Redirect(http.StatusSeeOther, "/routes")
// }

// func (h *Handler) DeleteDraftSpeedRequest(ctx *gin.Context) {
// 	// считываем значение из формы, которую мы добавим в наш шаблон
// 	strSpeedRequestId := ctx.Param("speed_request_id")
// 	speedRequestID, err := strconv.Atoi(strSpeedRequestId)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 	}
// 	// Вызов функции добавления чата в заявку
// 	err = h.Repository.DeleteDraftSpeedRequest(int(speedRequestID))
// 	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
// 		return
// 	}

// 	// после вызова сразу произойдет обновление страницы
// 	ctx.Redirect(http.StatusFound, "/routes")
// }
