package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"rip/internal/app/ds"
	"time"

	"github.com/sirupsen/logrus"
)

func (r *Repository) GetAllRoutes() ([]ds.Route, error) {
	var routes []ds.Route
	err := r.db.Find(&routes).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (r *Repository) GetRouteByID(route_id int) (*ds.Route, error) {
	query := "SELECT route_id, title, distance, description, delay, image_url FROM routes WHERE route_id = $1"

	// Создание курсора (строковый указатель)
	row := r.db.Raw(query, route_id).Row()

	// Создание объекта для хранения данных
	route := &ds.Route{}

	// Сканирование строки в структуру
	err := row.Scan(
		&route.RouteID,
		&route.Title,
		&route.Distance,
		&route.Description,
		&route.Delay,
		&route.ImageUrl,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Возвращаем nil, если записи нет
		}
		return nil, err
	}

	return route, nil
}

func (r *Repository) GetRoutesByDistance(minDistance, maxDistance int) ([]ds.Route, error) {
	var routes []ds.Route
	err := r.db.Find(&routes).Error
	if err != nil {
		return nil, err
	}

	var result []ds.Route
	for _, route := range routes {
		if route.Distance >= minDistance && route.Distance <= maxDistance {
			result = append(result, route)
		}
	}

	return result, nil
}

func (r *Repository) GetDraftCount() int64 {
	var requestID int
	var count int64
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.Request{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("request_id").First(&requestID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.RequestRoute{}).Where("request_id = ?", requestID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}

func (r *Repository) GetDraftRequest() (ds.Request, error) {
	var request ds.Request
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.Request{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("*").First(&request).Error
	if err != nil {
		logrus.Println("Error select darft request:", err)
	}

	return request, nil
}

func (r *Repository) GetRoutesByRequestID(requestID int) ([]ds.Route, error) {
	var routes []ds.Route

	err := r.db.Table("routes").Select("routes.*").
		Joins("INNER JOIN request_routes ON routes.route_id = request_routes.route_id").
		Where("request_routes.request_id = ?", requestID).
		Find(&routes).Error

	if err != nil {
		return nil, err
	}

	return routes, nil
}

func (r *Repository) GetRequestRoutes(requestID int) ([]ds.RequestRoute, error) {
	var requestRoutes []ds.RequestRoute

	err := r.db.
		Where("request_id = ?", requestID).
		Find(&requestRoutes).Error

	return requestRoutes, err
}

func (r *Repository) AddRouteToDraft(routeID int, userID int) error {
	var draft ds.Request
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&draft).Error

	if err != nil {
		// Создаем новый черновик если ещё нет
		draft = ds.Request{
			CreatorID:      userID,
			Status:         "черновик",
			CreationDate:   time.Now(),
			DepartureDate:  time.Date(2025, time.August, 5, 18, 0, 0, 0, time.UTC),
			ModeratorID:    2,
			FormationDate:  time.Time{},
			CompletionDate: time.Time{},
		}
		if err := r.db.Create(&draft).Error; err != nil {
			return err
		}
	}

	// Проверяем, не добавлен ли уже маршрут
	var existingRoute ds.RequestRoute
	err = r.db.Where("request_id = ? AND route_id = ?", draft.RequestID, routeID).First(&existingRoute).Error
	if err == nil {
		return fmt.Errorf("маршрут уже добавлен в заявку")
	}

	// Получаем информацию о маршруте для расчета скорости
	route, err := r.GetRouteByID(routeID)
	if err != nil {
		return fmt.Errorf("ошибка получения информации о маршруте: %w", err)
	}
	if route == nil {
		return fmt.Errorf("маршрут с ID %d не найден", routeID)
	}

	arrivalDate := time.Date(2025, time.August, 9, 15, 0, 0, 0, time.UTC)
	shipSpeed := calculateShipSpeed(
		draft.DepartureDate, // дата отправления
		arrivalDate,         // дата прибытия
		route.Delay,         // задержка в часах (предполагая, что Delay в часах)
		route.Distance,      // расстояние
	)

	// Добавляем маршрут в заявку с рассчитанной скоростью
	requestRoute := ds.RequestRoute{
		RequestID:   draft.RequestID,
		RouteID:     routeID,
		ArrivalDate: arrivalDate,
		ShipSpeed:   shipSpeed, // используем рассчитанную скорость
	}

	return r.db.Create(&requestRoute).Error
}

// func (r *Repository) DeleteDraftRequest(requestID int) error {
// 	err := r.db.Model(&ds.Request{}).Where("request_id = ?", requestID).UpdateColumn("status", "deleted").Error
// 	fmt.Println(requestID)
// 	if err != nil {
// 		return fmt.Errorf("ошибка при удалении чата с id %d: %w", requestID, err)
// 	}

// 	return nil
// }

func (r *Repository) DeleteDraftRequest(requestID int) error {
	// SQL запрос для обновления статуса заявки на "deleted"
	query := "UPDATE requests SET status = 'удалена' WHERE request_id = $1"

	// Выполняем SQL запрос
	result := r.db.Exec(query, requestID)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении заявки с id %d: %w", requestID, result.Error)
	}

	// Проверяем, была ли обновлена хотя бы одна строка
	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка с id %d не найдена", requestID)
	}

	return nil
}

func calculateShipSpeed(departureDate time.Time, arrivalDate time.Time, delay int, distance int) int {
	if departureDate.IsZero() || arrivalDate.IsZero() {
		return 0
	}
	travelDays := arrivalDate.Sub(departureDate).Hours() - float64(delay)
	speedKmh := distance / int(travelDays)
	speedKnots := float64(speedKmh) / 1.852
	return int(speedKnots)
}
