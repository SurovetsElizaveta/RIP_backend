package repository

import (
	"errors"
	"fmt"
	"mime/multipart"
	"rip/internal/app/ds"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *Repository) CreateRoute(route ds.Route) error {
	if route.Title == "" {
		return errors.New("название маршрута не может быть пустым")
	}
	if route.Distance <= 0 {
		return errors.New("дистанция должна быть положительным числом")
	}
	if route.Delay < 0 {
		return errors.New("задержка не может быть отрицательной")
	}

	if route.Status == "" {
		route.Status = "действует"
	}

	// Создаем запись в БД
	result := r.db.Create(route)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *Repository) UpdateRoute(route ds.Route) error {
	var existingRoute ds.Route
	if err := r.db.First(&existingRoute, route.RouteID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("маршрут не найден")
		}
		return err
	}

	if route.Title == "" {
		return errors.New("название маршрута не может быть пустым")
	}
	if route.Distance <= 0 {
		return errors.New("дистанция должна быть положительным числом")
	}

	result := r.db.Model(&ds.Route{}).Where("route_id = ?", route.RouteID).Updates(map[string]interface{}{
		"title":       route.Title,
		"distance":    route.Distance,
		"description": route.Description,
		"delay":       route.Delay,
		"image_url":   route.ImageURL,
		"status":      "действует",
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("маршрут не был обновлен")
	}

	return nil
}

func (r *Repository) DeleteRoute(routeId uint) error {
	var route ds.Route
	if err := r.db.First(&route, routeId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("маршрут не найден")
		}
		return err
	}

	// Если есть изображение - удаляем из Minio
	if route.ImageURL != "" {
		if err := r.deleteImageFromMinio(route.ImageURL); err != nil {
			// Логируем ошибку, но продолжаем удаление маршрута
			return err // или логируем и продолжаем
		}
	}

	// Удаляем связи в route_speed_requests (многие-ко-многим)
	if err := r.db.Where("route_id = ?", routeId).Delete(&ds.RouteSpeedRequest{}).Error; err != nil {
		return err
	}

	err := r.db.Model(&ds.Route{}).Where("route_id = ?", routeId).UpdateColumn("status", "удалён").Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении маршрута с id %d: %w", routeId, err)
	}

	return nil
}

func (r *Repository) GetAllRoutes() ([]ds.Route, error) {
	var routes []ds.Route
	err := r.db.Find(&routes).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (r *Repository) GetRouteByID(route_id uint) (ds.Route, error) {
	var route ds.Route
	err := r.db.Where("route_id = ? AND status = ?", route_id, "действует").First(&route).Error
	if err != nil {
		logrus.Println("Error select route by id:", err)
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

func (r *Repository) deleteImageFromMinio(imageURL string) error {
	if r.minio == nil {
		return errors.New("minio клиент не инициализирован")
	}

	return r.minio.DeleteImage(imageURL)
}

func (r *Repository) UploadRouteImage(routeId uint, file *multipart.FileHeader) error {
	// Проверяем существование маршрута
	var route ds.Route
	if err := r.db.First(&route, routeId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("маршрут не найден")
		}
		return err
	}

	// Если было старое изображение - удаляем его из Minio
	if route.ImageURL != "" {
		if err := r.minio.DeleteImage(route.ImageURL); err != nil {
			// Логируем ошибку, но продолжаем обновление
			logrus.Warnf("Не удалось удалить старое изображение: %v", err)
		}
	}

	// Загружаем новое изображение через Minio клиент
	imageURL, err := r.minio.UploadImage(file, int(routeId))
	if err != nil {
		return fmt.Errorf("ошибка загрузки изображения: %w", err)
	}

	// Обновляем URL изображения в БД
	result := r.db.Model(&route).Update("image_url", imageURL)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// #############################################################

// func (r *Repository) GetDraftCount() int64 {
// 	var speedRequest ds.SpeedRequest
// 	creatorID := 1

// 	// Ищем черновик
// 	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").First(&speedRequest).Error
// 	if err != nil {
// 		return 0 // Черновика нет
// 	}

// 	// Считаем маршруты в черновике
// 	var count int64
// 	err = r.db.Model(&ds.RouteSpeedRequest{}).Where("speed_request_id = ?", speedRequest.SpeedRequestID).Count(&count).Error
// 	if err != nil {
// 		logrus.Println("Error counting routes in draft:", err)
// 		return 0
// 	}

// 	return count
// }

// func (r *Repository) GetDraftSpeedRequest() (ds.SpeedRequest, error) {
// 	var speedRequest ds.SpeedRequest
// 	creatorID := 1
// 	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

// 	err := r.db.Model(&ds.SpeedRequest{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("*").First(&speedRequest).Error
// 	if err != nil {
// 		logrus.Println("Error select darft speed request:", err)
// 	}

// 	return speedRequest, nil
// }

// func (r *Repository) GetSpeedRequestByID(speedRequestID int) (ds.SpeedRequest, error) {
// 	var speedRequest ds.SpeedRequest
// 	err := r.db.First(&speedRequest, speedRequestID).Error
// 	if err != nil {
// 		return ds.SpeedRequest{}, err
// 	}
// 	return speedRequest, nil
// }

// func (r *Repository) GetRoutesBySpeedRequestID(speedRequestID int) ([]ds.Route, error) {
// 	var routes []ds.Route

// 	err := r.db.Table("routes").Select("routes.*").
// 		Joins("INNER JOIN route_speed_requests ON routes.route_id = route_speed_requests.route_id").
// 		Where("route_speed_requests.speed_request_id = ?", speedRequestID).
// 		Find(&routes).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	return routes, nil
// }

// func (r *Repository) GetSpeedRequestRoutes(speedRequestID int) ([]ds.RouteSpeedRequest, error) {
// 	var speedRequestRoutes []ds.RouteSpeedRequest

// 	err := r.db.
// 		Where("speed_request_id = ?", speedRequestID).
// 		Find(&speedRequestRoutes).Error

// 	return speedRequestRoutes, err
// }

// func (r *Repository) AddRouteToDraft(routeID uint, userID uint) error {
// 	var draft ds.SpeedRequest
// 	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&draft).Error

// 	if err != nil {
// 		// Создаем новый черновик если ещё нет
// 		draft = ds.SpeedRequest{
// 			CreatorID:      userID,
// 			Status:         "черновик",
// 			CreationDate:   time.Now(),
// 			DepartureDate:  time.Date(2025, time.August, 5, 18, 0, 0, 0, time.UTC),
// 			ModeratorID:    2,
// 			FormationDate:  time.Time{},
// 			CompletionDate: time.Time{},
// 		}
// 		if err := r.db.Create(&draft).Error; err != nil {
// 			return err
// 		}
// 	}

// 	// Проверяем, не добавлен ли уже маршрут
// 	var existingRoute ds.RouteSpeedRequest
// 	err = r.db.Where("speed_request_id = ? AND route_id = ?", draft.SpeedRequestID, routeID).First(&existingRoute).Error
// 	if err == nil {
// 		return fmt.Errorf("маршрут уже добавлен в заявку")
// 	}

// 	// Получаем информацию о маршруте для расчета скорости
// 	route, err := r.GetRouteByID(routeID)
// 	if err != nil {
// 		return fmt.Errorf("ошибка получения информации о маршруте: %w", err)
// 	}

// 	arrivalDate := time.Date(2025, time.August, 9, 15, 0, 0, 0, time.UTC)
// 	shipSpeed := calculateShipSpeed(
// 		draft.DepartureDate, // дата отправления
// 		arrivalDate,         // дата прибытия
// 		route.Delay,         // задержка в часах (предполагая, что Delay в часах)
// 		route.Distance,      // расстояние
// 	)

// 	// Добавляем маршрут в заявку с рассчитанной скоростью
// 	speedRequestRoute := ds.RouteSpeedRequest{
// 		SpeedRequestID: draft.SpeedRequestID,
// 		RouteID:        routeID,
// 		ArrivalDate:    arrivalDate,
// 		ShipSpeed:      shipSpeed, // используем рассчитанную скорость
// 	}

// 	return r.db.Create(&speedRequestRoute).Error
// }

// func (r *Repository) DeleteDraftSpeedRequest(speedRequestID int) error {
// 	// SQL запрос для обновления статуса заявки на "deleted"
// 	query := "UPDATE speed_requests SET status = 'удалена' WHERE speed_request_id = $1"

// 	// Выполняем SQL запрос
// 	result := r.db.Exec(query, speedRequestID)
// 	if result.Error != nil {
// 		return fmt.Errorf("ошибка при удалении заявки с id %d: %w", speedRequestID, result.Error)
// 	}

// 	// Проверяем, была ли обновлена хотя бы одна строка
// 	if result.RowsAffected == 0 {
// 		return fmt.Errorf("заявка с id %d не найдена", speedRequestID)
// 	}

// 	return nil
// }

// func calculateShipSpeed(departureDate time.Time, arrivalDate time.Time, delay int, distance int) int {
// 	if departureDate.IsZero() || arrivalDate.IsZero() {
// 		return 0
// 	}
// 	travelDays := arrivalDate.Sub(departureDate).Hours() - float64(delay)
// 	speedKmh := distance / int(travelDays)
// 	speedKnots := float64(speedKmh) / 1.852
// 	return int(speedKnots)
// }
