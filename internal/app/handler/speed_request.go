package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rip/internal/app/ds"
	"rip/internal/app/dto"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02.01.2006")
}

// GetDraftInfo godoc
// @Summary Get darft information
// @Description Get darft infomation. Returns speedrequestid nil and count 0 for guest.
// @Tags speedrequests
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Failture 500
// @Router /speedrequests/draft [get]
func (h *Handler) GetDraftInfo(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")

	if !exists {
		ctx.JSON(http.StatusOK, gin.H{
			"draft_id": nil,
			"count":    0,
		})
	}

	draft, err := h.Repository.GetDraftByUserID(currentUserID.(uint))
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

// GetAllSpeedRequests godoc
// @Summary Get speed requests list
// @Description Get speed requests list. For authentificated users only.
// @Tags speedrequests
// @Produce json
// @Security BearerAuth
// @Param date_from query string false "Date From" Format(date)
// @Param date_to query string false "Date To" Rormat(date)
// @Param status query string false "Status"
// @Success 200 {array} dto.SpeedRequest "response"
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /speedrequests [get]
func (h *Handler) GetAllSpeedRequests(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

	isModerator, _ := ctx.Get("is_moderator")

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

	speedRequests, err := h.Repository.GetSpeedRequestsWithFilters(
		status,
		dateFrom,
		dateTo,
		currentUserID.(uint),
		isModerator.(bool),
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	response := make([]dto.SpeedRequest, len(speedRequests))
	for i, sr := range speedRequests {
		resultsCount, err := h.Repository.GetSpeedRequestResultsCount(sr.SpeedRequestID)
		if err != nil {
			resultsCount = 0 // В случае ошибки устанавливаем 0
		}

		response[i] = dto.SpeedRequest{
			SpeedRequestID: sr.SpeedRequestID,
			DepartureDate:  formatDate(sr.DepartureDate),
			CreationDate:   formatDate(sr.CreationDate),
			FormationDate:  formatDate(sr.FormationDate),
			CompletionDate: formatDate(sr.CompletionDate),
			Status:         sr.Status,
			CreatorLogin:   sr.Creator.Login,
			ModeratorLogin: getModeratorLogin(sr.Moderator),
			ResultsCount:   resultsCount,
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// GetSpeedRequestByID godoc
// @Summary Get speed request by ID
// @Description Get speed request by ID. For authentificated users only. Client can see only their speed requests
// @Tags speedrequests
// @Produce json
// @Security BearerAuth
// @Param speed_request_id path int true "Speed Request ID"
// @Success 200 {object} dto.SpeedRequestDetailedResponse "response"
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /speedrequests/{speed_request_id} [get]
func (h *Handler) GetSpeedRequestByID(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")

	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

	speedRequestID, err := strconv.ParseUint(ctx.Param("speed_request_id"), 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	speedRequest, routes, err := h.Repository.GetSpeedRequestWithRoutes(uint(speedRequestID), currentUserID.(uint))
	if err != nil {
		if err.Error() == "заявка не найдена" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if err.Error() == "пользователь не найден" {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else if err.Error() == "доступ запрещен" {
			h.errorHandler(ctx, http.StatusForbidden, err)
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

// UpdateSpeedRequest godoc
// @Summary Update speed request
// @Description Update speed request. For authentificated users only.
// @Tags speedrequests
// @Produce json
// @Security BearerAuth
// @Param speed_request_id path int true "Speed Request ID"
// @Success 200 {object} dto.UpdateSpeedRequest "request"
// @Failture 404
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /speedrequests/{speed_request_id} [put]
func (h *Handler) UpdateSpeedRequest(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")

	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}

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

	if err := h.Repository.UpdateSpeedRequest(uint(speedRequestID), currentUserID.(uint), request); err != nil {
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

// SubmitSpeedRequest godoc
// @Summary Sumbit speed request
// @Description Submit speed request. For authentificated users only.
// @Tags speedrequests
// @Produce json
// @Security BearerAuth
// @Param speed_request_id path int true "Speed Request ID"
// @Success 200 {object} dto.SpeedRequestDetailedResponse "response"
// @Failture 404
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /speedrequests/{speed_request_id}/submit [put]
func (h *Handler) SubmitSpeedRequest(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")

	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}
	speedRequestID, err := strconv.Atoi(ctx.Param("speed_request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	if err := h.Repository.SubmitSpeedRequest(uint(speedRequestID), currentUserID.(uint)); err != nil {
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

// CompleteSpeedRequest godoc
// @Summary Complete speed request
// @Description Complete speed requests. For moderators only.
// @Tags speedrequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param speed_request_id path int true "Speed Request ID"
// @Success 200 {object} object
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /speedrequests/{speed_request_id}/complete [put]
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
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("пользователь не аутентифицирован"))
		return
	}

	// Если статус "завершена", запускаем асинхронные расчеты для каждого маршрута
	if request.Status == "завершена" {
		speedRequest, routes, err := h.Repository.GetSpeedRequestWithRoutes(uint(speedRequestID), userID.(uint))
		if err != nil {
			if err.Error() == "заявка не найдена" {
				h.errorHandler(ctx, http.StatusNotFound, err)
			} else {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
			}
			return
		}

		// Отправляем запросы в асинхронный сервис для каждого маршрута
		for _, route := range routes {
			if err := h.AsyncService.RequestSpeedCalculation(
				speedRequest.SpeedRequestID,
				route.RouteID,
				float64(route.Route.Distance),
				float64(route.Route.Delay),
				speedRequest.DepartureDate, // Дата отправления из заявки
				route.ArrivalDate,          // Дата прибытия из маршрута
			); err != nil {
				// Логируем ошибку, но продолжаем обработку других маршрутов
				logrus.Errorf("Ошибка отправки запроса для маршрута %d: %v", route.RouteID, err)
			}
		}
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

// DeleteSpeedRequest godoc
// @Summary Delete speed request
// @Description Delete speed request. For authentificated users only.
// @Tags speedrequests
// @Produce json
// @Security BearerAuth
// @Param speed_request_id path int true "Speed Request ID"
// @Success 200 {object} object
// @Failture 403
// @Failture 400
// @Failture 500
// @Router /speedrequests/{speed_request_id} [delete]
func (h *Handler) DeleteSpeedRequest(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")

	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("user not authenticated"))
		return
	}
	speedRequestID, err := strconv.Atoi(ctx.Param("speed_request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	if err := h.Repository.DeleteSpeedRequest(uint(speedRequestID), currentUserID.(uint)); err != nil {
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

// ReceiveAsyncResult godoc
// @Summary Receive async calculation result
// @Description Receive result from async service for speed calculation
// @Tags speedrequests
// @Accept json
// @Produce json
// @Param request body object true "Async result data"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 500 {object} object
// @Router /async/result [post]
func (h *Handler) ReceiveAsyncResult(ctx *gin.Context) {
	// Логируем входящий запрос
	logrus.Infof("Получен запрос от Django: метод=%s, путь=%s",
		ctx.Request.Method, ctx.Request.URL.Path)

	// Проверяем метод
	if ctx.Request.Method != "POST" {
		h.errorHandler(ctx, http.StatusMethodNotAllowed,
			fmt.Errorf("метод %s не разрешен, ожидается POST", ctx.Request.Method))
		return
	}

	// Проверяем Content-Type
	contentType := ctx.GetHeader("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("неверный Content-Type: %s, ожидается application/json", contentType))
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("ошибка чтения тела запроса: %v", err))
		return
	}

	if len(body) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("пустое тело запроса"))
		return
	}

	logrus.Infof("Тело запроса от Django: %s", string(body))

	var request struct {
		Success         bool    `json:"success"`
		CalculatedSpeed float64 `json:"calculated_speed"`
		SpeedRequestID  uint    `json:"speed_request_id" binding:"required"`
		RouteID         uint    `json:"route_id" binding:"required"`
		AuthToken       string  `json:"auth_token" binding:"required"`
		Message         string  `json:"message"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		logrus.Errorf("Ошибка парсинга JSON: %v, тело: %s", err, string(body))
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("неверный формат JSON: %v", err))
		return
	}

	// Проверка токена
	expectedToken := h.Config.AsyncService.Token
	if request.AuthToken != expectedToken {
		h.errorHandler(ctx, http.StatusUnauthorized,
			fmt.Errorf("неверный токен авторизации"))
		return
	}

	logrus.Infof("Обработка результата: success=%v, speed=%f, speed_request_id=%d, route_id=%d",
		request.Success, request.CalculatedSpeed, request.SpeedRequestID, request.RouteID)

	// Обновляем скорость
	var shipSpeed int
	if request.Success && request.CalculatedSpeed > 0 {
		shipSpeed = int(request.CalculatedSpeed)
	} else {
		shipSpeed = 0
	}

	if err := h.Repository.UpdateRouteSpeedRequestShipSpeed(
		request.SpeedRequestID,
		request.RouteID,
		shipSpeed,
	); err != nil {
		logrus.Errorf("Ошибка обновления скорости: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":          "Результат успешно обработан",
		"speed_request_id": request.SpeedRequestID,
		"route_id":         request.RouteID,
		"ship_speed":       shipSpeed,
	})
}
