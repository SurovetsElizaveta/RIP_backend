package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"rip/internal/app/ds"
	"rip/internal/app/dto"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetDraftInfo(ctx *gin.Context) {
	currentUserID := 1

	draft, err := h.Repository.GetDraftByUserID(uint(currentUserID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if draft == (ds.SpeedRequest{}) {
		ctx.JSON(http.StatusOK, gin.H{
			"draft_id": nil,
			"count":    0,
		})
		return
	}

	count, err := h.Repository.GetSpeedRequestRoutesCount(draft.SpeedRequestID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"draft_id": draft.SpeedRequestID,
		"count":    count,
	})
}

func (h *Handler) GetAllSpeedRequests(ctx *gin.Context) {
	// Получаем параметры запроса
	status := ctx.Query("status")
	dateFromStr := ctx.Query("date_from")
	dateToStr := ctx.Query("date_to")

	var dateFrom, dateTo *time.Time
	var err error

	// Парсим date_from
	if dateFromStr != "" {
		parsedDateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат date_from: %v", err))
			return
		}
		dateFrom = &parsedDateFrom
	}

	// Парсим date_to
	if dateToStr != "" {
		parsedDateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат date_to: %v", err))
			return
		}
		dateTo = &parsedDateTo
	}

	// Валидация дат
	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("date_from не может быть после date_to"))
		return
	}

	// Вызываем репозиторий с простыми параметрами
	speedRequests, err := h.Repository.GetSpeedRequestsWithFilters(status, dateFrom, dateTo)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	response := make([]dto.SpeedRequest, len(speedRequests))
	for i, sr := range speedRequests {
		response[i] = dto.SpeedRequest{
			SpeedRequestID: sr.SpeedRequestID,
			DepartureDate:  sr.DepartureDate,
			CreationDate:   sr.CreationDate,
			FormationDate:  sr.FormationDate,
			CompletionDate: sr.CompletionDate,
			Status:         sr.Status,
			CreatorLogin:   sr.Creator.Login,
			ModeratorLogin: getModeratorLogin(sr.Moderator),
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) GetSpeedRequestByID(ctx *gin.Context) {
	speedRequestID, err := strconv.ParseUint(ctx.Param("speed_request_id"), 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	speedRequest, routes, err := h.Repository.GetSpeedRequestWithRoutes(uint(speedRequestID))
	if err != nil {
		if err.Error() == "заявка не найдена" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	if speedRequest.Status == ds.StatusDeleted {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("заявка не найдена"))
		return
	}

	// Формируем ответ вручную без объединяющей структуры
	response := gin.H{
		"id":              speedRequest.SpeedRequestID,
		"departure_date":  speedRequest.DepartureDate,
		"creation_date":   speedRequest.CreationDate,
		"formation_date":  speedRequest.FormationDate,
		"completion_date": speedRequest.CompletionDate,
		"status":          speedRequest.Status,
		"creator_login":   speedRequest.Creator.Login,
		"moderator_login": getModeratorLogin(speedRequest.Moderator),
		"routes":          make([]gin.H, len(routes)),
	}

	// Добавляем маршруты в ответ
	for i, route := range routes {
		response["routes"].([]gin.H)[i] = gin.H{
			"route_id":     route.RouteID,
			"title":        route.Route.Title,
			"distance":     route.Route.Distance,
			"description":  route.Route.Description,
			"image_url":    route.Route.ImageURL,
			"arrival_date": route.ArrivalDate,
			"ship_speed":   route.ShipSpeed,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) UpdateSpeedRequest(ctx *gin.Context) {
	speedRequestID, err := strconv.Atoi(ctx.Param("speed_request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	var request dto.UpdateSpeedRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	currentUserID := 1

	if err := h.Repository.UpdateSpeedRequest(uint(speedRequestID), uint(currentUserID), request); err != nil {
		if err.Error() == "заявка не найдена" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if err.Error() == "доступ запрещен" {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	// ctx.JSON(http.StatusOK, gin.H{
	// 	"message": "Заявка успешно обновлена",
	// })
	ctx.JSON(http.StatusOK, request)
}

func (h *Handler) SubmitSpeedRequest(ctx *gin.Context) {
	speedRequestID, err := strconv.Atoi(ctx.Param("speed_request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	currentUserID := 1

	if err := h.Repository.SubmitSpeedRequest(uint(speedRequestID), uint(currentUserID)); err != nil {
		if err.Error() == "заявка не найдена" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if err.Error() == "доступ запрещен" {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else if err.Error() == "не все обязательные поля заполнены" {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Заявка успешно сформирована",
	})
}

func (h *Handler) CompleteSpeedRequest(ctx *gin.Context) {
	speedRequestID, err := strconv.Atoi(ctx.Param("speed_request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	var request struct {
		Status string `json:"status" binding:"required,oneof=завершена отклонена"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	moderatorID := 2

	if err := h.Repository.CompleteSpeedRequest(uint(speedRequestID), uint(moderatorID), request.Status); err != nil {
		if err.Error() == "заявка не найдена" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if err.Error() == "доступ запрещен" {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else if err.Error() == "неверный статус" {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Заявка успешно %s", request.Status),
	})
}

func (h *Handler) DeleteSpeedRequest(ctx *gin.Context) {
	speedRequestID, err := strconv.Atoi(ctx.Param("speed_request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	currentUserID := 1

	if err := h.Repository.DeleteSpeedRequest(uint(speedRequestID), uint(currentUserID)); err != nil {
		if err.Error() == "заявка не найдена" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if err.Error() == "доступ запрещен" {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Заявка успешно удалена",
	})
}

func getModeratorLogin(moderator ds.User) string {
	if moderator.UserID == 0 {
		return ""
	}
	return moderator.Login
}
