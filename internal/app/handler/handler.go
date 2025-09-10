package handler

import (
	"Lab1/internal/app/repository"
	"net/http"
	"strconv"

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

func (h *Handler) GetRoutes(ctx *gin.Context) {
	var routes []repository.Route
	var err error

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		routes, err = h.Repository.GetRoutes()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		routes, err = h.Repository.GetRoutesByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"routes": routes,
		"query":  searchQuery, // передаем введенный запрос обратно на страницу
		// в ином случае оно будет очищаться при нажатии на кнопку
	})
}

func (h *Handler) GetRoute(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	route, err := h.Repository.GetRoute(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "route.html", gin.H{
		"route": route,
	})
}

func (h *Handler) GetRequests(ctx *gin.Context) {
	requestRoutes, err := h.Repository.GetRequestDraft()
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	routes, err := h.Repository.GetRoutes()
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Создаем map для быстрого поиска заказов по ID
	routesMap := make(map[int]repository.Route)
	for _, route := range routes {
		routesMap[route.ID] = route
	}

	ctx.HTML(http.StatusOK, "request.html", gin.H{
		"requestRoutes": requestRoutes,
		"routesMap":     routesMap, // Передаем map вместо массива
	})
}
