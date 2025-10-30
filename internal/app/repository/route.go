package repository

import (
	"errors"
	"fmt"
	"mime/multipart"
	"rip/internal/app/ds"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *Repository) GetAllRoutes() ([]ds.Route, error) {
	var routes []ds.Route
	err := r.db.Where("status = ?", "действует").Find(&routes).Error
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (r *Repository) GetRoutesByDistance(minDistance, maxDistance int) ([]ds.Route, error) {
	var routes []ds.Route

	err := r.db.Where("status = ? AND distance >= ? AND distance <= ?", "действует", minDistance, maxDistance).Find(&routes).Error
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

func (r *Repository) CreateRoute(route *ds.Route) error {
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

	if route.ImageURL != "" {
		if err := r.deleteImageFromMinio(route.ImageURL); err != nil {
			return err
		}
	}

	if err := r.db.Where("route_id = ?", routeId).Delete(&ds.RouteSpeedRequest{}).Error; err != nil {
		return err
	}

	err := r.db.Model(&ds.Route{}).Where("route_id = ?", routeId).UpdateColumn("status", "удалён").Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении маршрута с id %d: %w", routeId, err)
	}

	return nil
}

func (r *Repository) deleteImageFromMinio(imageURL string) error {
	if r.minio == nil {
		return errors.New("minio клиент не инициализирован")
	}

	return r.minio.DeleteImage(imageURL)
}

func (r *Repository) UploadRouteImage(routeId uint, file *multipart.FileHeader) error {
	var route ds.Route
	if err := r.db.First(&route, routeId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("маршрут не найден")
		}
		return err
	}

	if route.ImageURL != "" {
		if err := r.minio.DeleteImage(route.ImageURL); err != nil {
			logrus.Warnf("Не удалось удалить старое изображение: %v", err)
		}
	}

	imageURL, err := r.minio.UploadImage(file, int(routeId))
	if err != nil {
		return fmt.Errorf("ошибка загрузки изображения: %w", err)
	}

	result := r.db.Model(&route).Update("image_url", imageURL)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *Repository) AddRouteToDraft(routeID uint, userID uint) (uint, error) {
	var route ds.Route
	if err := r.db.First(&route, routeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("маршрут не найден")
		}
		return 0, err
	}

	draft, err := r.getOrCreateDraft(userID)
	if err != nil {
		return 0, err
	}

	var existingRoute ds.RouteSpeedRequest
	err = r.db.Where("speed_request_id = ? AND route_id = ?", draft.SpeedRequestID, routeID).First(&existingRoute).Error
	if err == nil {
		return 0, errors.New("маршрут уже добавлен в заявку")
	}

	routeSpeedRequest := ds.RouteSpeedRequest{
		SpeedRequestID: draft.SpeedRequestID,
		RouteID:        routeID,
		ArrivalDate:    time.Time{},
		ShipSpeed:      0,
	}

	if err := r.db.Create(&routeSpeedRequest).Error; err != nil {
		return 0, fmt.Errorf("ошибка добавления маршрута в заявку: %w", err)
	}

	logrus.Infof("Маршрут %d добавлен в заявку-черновик %d", routeID, draft.SpeedRequestID)
	return draft.SpeedRequestID, nil
}

func (r *Repository) getOrCreateDraft(userID uint) (*ds.SpeedRequest, error) {
	var draft ds.SpeedRequest

	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&draft).Error
	if err == nil {
		return &draft, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newDraft := ds.SpeedRequest{
		CreatorID:    userID,
		Status:       ds.StatusDraft,
		CreationDate: time.Now(),
	}

	if err := r.db.Create(&newDraft).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания черновика: %w", err)
	}

	logrus.Infof("Создан новый черновик заявки ID: %d для пользователя %d", newDraft.SpeedRequestID, userID)
	return &newDraft, nil
}
