package repository

import (
	"errors"
	"fmt"
	"time"

	"rip/internal/app/ds"
	"rip/internal/app/dto"

	"gorm.io/gorm"
)

func (r *Repository) GetDraftByUserID(userID uint) (ds.SpeedRequest, error) {
	var draft ds.SpeedRequest
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&draft).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.SpeedRequest{}, nil
		}
		return ds.SpeedRequest{}, err
	}
	return draft, nil
}

func (r *Repository) GetSpeedRequestRoutesCount(speedRequestID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.RouteSpeedRequest{}).Where("speed_request_id = ?", speedRequestID).Count(&count).Error
	return count, err
}

func (r *Repository) GetSpeedRequestsWithFilters(status string, dateFrom, dateTo *time.Time) ([]ds.SpeedRequest, error) {
	var speedRequests []ds.SpeedRequest

	query := r.db.Model(&ds.SpeedRequest{}).
		Preload("Creator").
		Preload("Moderator").
		Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if dateFrom != nil {
		startOfDay := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, dateFrom.Location())
		query = query.Where("formation_date >= ?", startOfDay)
	}

	if dateTo != nil {
		endOfDay := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 0, dateTo.Location())
		query = query.Where("formation_date <= ?", endOfDay)
	}

	query = query.Order("formation_date DESC")

	err := query.Find(&speedRequests).Error
	return speedRequests, err
}

func (r *Repository) GetSpeedRequestWithRoutes(speedRequestID uint) (ds.SpeedRequest, []ds.RouteSpeedRequest, error) {
	// Получаем заявку
	var speedRequest ds.SpeedRequest
	err := r.db.
		Preload("Creator").
		Preload("Moderator").
		First(&speedRequest, speedRequestID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.SpeedRequest{}, nil, errors.New("заявка не найдена")
		}
		return ds.SpeedRequest{}, nil, err
	}

	// Получаем маршруты заявки отдельным запросом
	var routes []ds.RouteSpeedRequest
	err = r.db.
		Preload("Route").
		Where("speed_request_id = ?", speedRequestID).
		Find(&routes).Error

	if err != nil {
		return ds.SpeedRequest{}, nil, err
	}

	return speedRequest, routes, nil
}

func (r *Repository) UpdateSpeedRequest(speedRequestID uint, userID uint, request dto.UpdateSpeedRequest) error {
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

	updates := make(map[string]interface{})
	if request.DepartureDate != nil {
		updates["departure_date"] = *request.DepartureDate
	}

	if len(updates) == 0 {
		return errors.New("нет полей для обновления")
	}

	return r.db.Model(&speedRequest).Updates(updates).Error
}

func (r *Repository) SubmitSpeedRequest(speedRequestID uint, userID uint) error {
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
		return errors.New("можно формировать только черновики")
	}

	if err := r.validateSpeedRequestSubmission(speedRequest); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"status":         ds.StatusSubmitted,
		"formation_date": time.Now(),
	}

	return r.db.Model(&speedRequest).Updates(updates).Error
}

func (r *Repository) validateSpeedRequestSubmission(speedRequest ds.SpeedRequest) error {
	if speedRequest.DepartureDate.IsZero() {
		return errors.New("дата отправления обязательна")
	}

	count, err := r.GetSpeedRequestRoutesCount(speedRequest.SpeedRequestID)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("заявка должна содержать хотя бы одну услугу")
	}

	return nil
}

func (r *Repository) CompleteSpeedRequest(speedRequestID uint, moderatorID uint, status string) error {
	var speedRequest ds.SpeedRequest
	if err := r.db.First(&speedRequest, speedRequestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("заявка не найдена")
		}
		return err
	}

	if speedRequest.Status != ds.StatusSubmitted {
		return errors.New("можно завершать/отклонять только сформированные заявки")
	}

	if status != ds.StatusCompleted && status != ds.StatusRejected {
		return errors.New("неверный статус")
	}

	if status == ds.StatusCompleted {
		if err := r.calculateShipSpeedRequest(speedRequest); err != nil {
			return fmt.Errorf("ошибка расчета бизнес-логики: %w", err)
		}
	}

	updates := map[string]interface{}{
		"status":          status,
		"moderator_id":    moderatorID,
		"completion_date": time.Now(),
	}

	return r.db.Model(&speedRequest).Updates(updates).Error
}

func (r *Repository) DeleteSpeedRequest(speedRequestID uint, userID uint) error {
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

	updates := map[string]interface{}{
		"status": "удалена",
	}

	return r.db.Model(&speedRequest).Updates(updates).Error
}

func (r *Repository) calculateShipSpeedRequest(speedRequest ds.SpeedRequest) error {
	var routeSpeedRequests []ds.RouteSpeedRequest
	if err := r.db.
		Preload("Route").
		Where("speed_request_id = ?", speedRequest.SpeedRequestID).
		Find(&routeSpeedRequests).Error; err != nil {
		return fmt.Errorf("ошибка получения данных маршрутов: %w", err)
	}

	for i := range routeSpeedRequests {
		routeSpeedRequest := &routeSpeedRequests[i]

		shipSpeed := r.calculateShipSpeed(
			speedRequest.DepartureDate,
			routeSpeedRequest.ArrivalDate,
			routeSpeedRequest.Route.Delay,
			routeSpeedRequest.Route.Distance,
		)

		if err := r.db.Model(routeSpeedRequest).Update("ship_speed", shipSpeed).Error; err != nil {
			return fmt.Errorf("ошибка обновления скорости судна: %w", err)
		}
	}

	return nil
}

func (r *Repository) calculateShipSpeed(departureDate, arrivalDate time.Time, delay, distance int) int {
	if departureDate.IsZero() || arrivalDate.IsZero() {
		return 0
	}
	travelHours := arrivalDate.Sub(departureDate).Hours() - float64(delay)
	if travelHours <= 0 {
		return 0
	}
	speedKmh := float64(distance) / travelHours
	speedKnots := speedKmh / 1.852
	return int(speedKnots)
}
