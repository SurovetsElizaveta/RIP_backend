package repository

import (
	"errors"

	"rip/internal/app/ds"

	"gorm.io/gorm"
)

func (r *Repository) RemoveRouteFromSpeedRequest(speedRequestID uint, routeID uint) error {
	result := r.db.Where("speed_request_id = ? AND route_id = ?", speedRequestID, routeID).Delete(&ds.RouteSpeedRequest{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("услуга не найдена в заявке")
	}
	return nil
}

func (r *Repository) UpdateRouteSpeedRequest(speedRequestID uint, routeID uint, updates map[string]interface{}) error {
	var routeSpeedRequest ds.RouteSpeedRequest
	if err := r.db.Where("speed_request_id = ? AND route_id = ?", speedRequestID, routeID).First(&routeSpeedRequest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("связь не найдена")
		}
		return err
	}

	return r.db.Model(&routeSpeedRequest).Updates(updates).Error
}

func (r *Repository) ValidateSpeedRequestAccess(speedRequestID uint, userID uint) error {
	var speedRequest ds.SpeedRequest
	if err := r.db.First(&speedRequest, speedRequestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("заявка не найдена")
		}
		return err
	}

	if speedRequest.CreatorID != userID {
		return errors.New("доступ запрещен")
	}

	if speedRequest.Status != ds.StatusDraft {
		return errors.New("можно редактировать только черновики")
	}

	return nil
}

func (r *Repository) GetRouteSpeedRequest(speedRequestID uint, routeID uint) (*ds.RouteSpeedRequest, error) {
	var routeSpeedRequest ds.RouteSpeedRequest
	err := r.db.
		Preload("Route").
		Where("speed_request_id = ? AND route_id = ?", speedRequestID, routeID).
		First(&routeSpeedRequest).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("связь не найдена")
		}
		return nil, err
	}

	return &routeSpeedRequest, nil
}
