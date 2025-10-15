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

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02.01.2006")
}

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
	status := ctx.Query("status")
	dateFromStr := ctx.Query("date_from")
	dateToStr := ctx.Query("date_to")

	var dateFrom, dateTo *time.Time
	var err error

	if dateFromStr != "" {
		parsedDateFrom, err := time.Parse("02.01.2006", dateFromStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат date_from: %v", err))
			return
		}
		dateFrom = &parsedDateFrom
	}

	if dateToStr != "" {
		parsedDateTo, err := time.Parse("02.01.2006", dateToStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат date_to: %v", err))
			return
		}
		dateTo = &parsedDateTo
	}

	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("date_from не может быть после date_to"))
		return
	}

	speedRequests, err := h.Repository.GetSpeedRequestsWithFilters(status, dateFrom, dateTo)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	response := make([]dto.SpeedRequest, len(speedRequests))
	for i, sr := range speedRequests {
		response[i] = dto.SpeedRequest{
			SpeedRequestID: sr.SpeedRequestID,
			DepartureDate:  formatDate(sr.DepartureDate),
			CreationDate:   formatDate(sr.CreationDate),
			FormationDate:  formatDate(sr.FormationDate),
			CompletionDate: formatDate(sr.CompletionDate),
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

	routeReq := make([]dto.RouteSpeedRequestInfo, len(routes))
	routeInfos := make([]dto.RouteInfo, len(routes))

	speedReq := dto.SpeedRequest{
		SpeedRequestID: speedRequest.SpeedRequestID,
		DepartureDate:  formatDate(speedRequest.DepartureDate),
		CreationDate:   formatDate(speedRequest.CreationDate),
		FormationDate:  formatDate(speedRequest.FormationDate),
		CompletionDate: formatDate(speedRequest.CompletionDate),
		CreatorLogin:   speedRequest.Creator.Login,
		ModeratorLogin: speedRequest.Moderator.Login,
		Status:         speedRequest.Status,
	}

	var result int

	for i, route := range routes {
		routeReq[i] = dto.RouteSpeedRequestInfo{
			ArrivalDate: formatDate(route.ArrivalDate),
			ShipSpeed:   route.ShipSpeed,
		}

		if route.ShipSpeed != 0 {
			result += 1
		}

		routeInfos[i] = dto.RouteInfo{
			RouteID:     route.Route.RouteID,
			Title:       route.Route.Title,
			Distance:    route.Route.Distance,
			Description: route.Route.Description,
			ImageURL:    route.Route.ImageURL,
			Status:      route.Route.Status,
		}
	}

	response := dto.SpeedRequestDetailedResponse{
		SpeedRequest: speedReq,
		RouteReq:     routeReq,
		Routes:       routeInfos,
		Result:       result,
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

	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("пользователь не аутентифицирован"))
		return
	}

	if err := h.Repository.CompleteSpeedRequest(uint(speedRequestID), userID.(uint), request.Status); err != nil {
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
