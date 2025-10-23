package handler

import (
	"fmt"
	"net/http"
	"time"

	"rip/internal/app/dto"

	"github.com/gin-gonic/gin"
)

// RemoveRouteSpeedRequest godoc
// @Summary Remove route speed request
// @Description Remove route from speed request. For authentificated users only.
// @Tags routespeedrequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 403
// @Failure 400
// @Failure 500
// @Router /routespeedrequests [delete]
func (h *Handler) RemoveRouteSpeedRequest(ctx *gin.Context) {
	var request dto.RemoveRouteSpeedRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("пользователь не аутентифицирован"))
		return
	}

	if err := h.Repository.ValidateSpeedRequestAccess(uint(request.SpeedRequestID), userID.(uint)); err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}

	if err := h.Repository.RemoveRouteFromSpeedRequest(uint(request.SpeedRequestID), uint(request.RouteID)); err != nil {
		if err.Error() == "маршрут не найден в заявке" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Маршрут успешно удален из заявки",
	})
}

// UpdateRouteSpeedRequest godoc
// @Summary Update route speed request
// @Description Upadet field arrival date in route speed request. For authentificated users only.
// @Tags routespeedrequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 403
// @Failure 400
// @Failure 500
// @Router /routespeedrequests [put]
func (h *Handler) UpdateRouteSpeedRequest(ctx *gin.Context) {
	var request dto.UpdateRouteSpeedRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("пользователь не аутентифицирован"))
		return
	}

	if err := h.Repository.ValidateSpeedRequestAccess(uint(request.SpeedRequestID), userID.(uint)); err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}

	updates := make(map[string]interface{})
	parsedArrivalDate, err := time.Parse("02.01.2006", request.ArrivalDate)
	if err != nil {
		return
	}
	if parsedArrivalDate != (time.Time{}) {
		updates["arrival_date"] = parsedArrivalDate
	}

	if len(updates) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("нет полей для обновления"))
		return
	}

	if err := h.Repository.UpdateRouteSpeedRequest(uint(request.SpeedRequestID), uint(request.RouteID), updates); err != nil {
		if err.Error() == "связь не найдена" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Данные успешно обновлены",
	})
}
