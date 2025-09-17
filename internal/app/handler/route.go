package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"rip/internal/app/ds"

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

	// Преобразуем обратно в строки для отображения в форме
	minDistanceValue := ctx.Query("min_distance")
	maxDistanceValue := ctx.Query("max_distance")

	draftCount := h.Repository.GetDraftCount()

	ctx.HTML(http.StatusOK, "routes.page.tmpl", gin.H{
		"routes":      routes,
		"minDistance": minDistanceValue,
		"maxDistance": maxDistanceValue,
		"draftCount":  draftCount,
	})
}

func (h *Handler) GetRouteById(ctx *gin.Context) {
	strId := ctx.Param("route_id")
	route_id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	route, err := h.Repository.GetRouteByID(route_id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "route.page.tmpl", route)
}

func (h *Handler) GetDraftRequest(ctx *gin.Context) {
	request, err := h.Repository.GetDraftRequest()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	routes, err := h.Repository.GetRoutesByRequestID(request.RequestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	requestRoutes, err := h.Repository.GetRequestRoutes(request.RequestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	requestRoutesMap := make(map[int]ds.RequestRoute)
	for _, rr := range requestRoutes {
		requestRoutesMap[rr.RouteID] = rr
	}

	ctx.HTML(http.StatusOK, "draft.page.tmpl", gin.H{
		"request":          request,
		"routes":           routes,
		"requestRoutesMap": requestRoutesMap,
	})
}

func (h *Handler) AddToDraft(ctx *gin.Context) {
	routeIDStr := ctx.Param("route_id")
	routeID, err := strconv.Atoi(routeIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID маршрута"))
		return
	}

	// TODO: Получить userID из сессии (пока хардкод)
	userID := 1

	err = h.Repository.AddRouteToDraft(routeID, userID)
	if err != nil {
		// Перенаправляем с сообщением об ошибке
		ctx.Redirect(http.StatusSeeOther, "/routes?error="+url.QueryEscape(err.Error()))
		return
	}

	// Перенаправляем с сообщением об успехе
	ctx.Redirect(http.StatusSeeOther, "/routes")
}

func (h *Handler) DeleteDraftRequest(ctx *gin.Context) {
	// считываем значение из формы, которую мы добавим в наш шаблон
	strRequestId := ctx.Param("request_id")
	requestID, err := strconv.Atoi(strRequestId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	// Вызов функции добавления чата в заявку
	err = h.Repository.DeleteDraftRequest(int(requestID))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return
	}

	// после вызова сразу произойдет обновление страницы
	ctx.Redirect(http.StatusFound, "/routes")
}

// func (h *Handler) DeleteChat(ctx *gin.Context) {
// 	// считываем значение из формы, которую мы добавим в наш шаблон
// 	strId := ctx.PostForm("chat_id")
// 	id, err := strconv.Atoi(strId)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 	}
// 	// Вызов функции добавления чата в заявку
// 	err = h.Repository.DeleteChat(uint(id))
// 	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
// 		return
// 	}

// 	// после вызова сразу произойдет обновление страницы
// 	ctx.Redirect(http.StatusFound, "/chats")
// }
