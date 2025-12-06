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

func (r *Repository) GetSpeedRequestsWithFilters(
	status string,
	dateFrom, dateTo *time.Time,
	userID uint,
	isModerator bool,
) ([]ds.SpeedRequest, error) {
	var speedRequests []ds.SpeedRequest

	query := r.db.Model(&ds.SpeedRequest{}).
		Preload("Creator").
		Preload("Moderator")

	if !isModerator {
		query = query.Where("creator_id = ?", userID).Where("status NOT IN (?, ?)", "черновик", "удалена")
	}

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

func (r *Repository) GetSpeedRequestWithRoutes(speedRequestID uint, userID uint) (ds.SpeedRequest, []ds.RouteSpeedRequest, error) {
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

	var user ds.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return ds.SpeedRequest{}, nil, errors.New("пользователь не найден")
	}

	if speedRequest.CreatorID != userID && !user.IsModerator {
		return ds.SpeedRequest{}, nil, errors.New("доступ запрещен")
	}

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
	parsedDepartureDate, err := time.Parse("02.01.2006", request.DepartureDate)
	if err != nil {
		return err
	}
	if parsedDepartureDate != (time.Time{}) {
		updates["departure_date"] = parsedDepartureDate
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

	if err := r.validateSpeedRequestSubmission(speedRequest.SpeedRequestID, userID); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"status":         ds.StatusSubmitted,
		"formation_date": time.Now(),
	}

	return r.db.Model(&speedRequest).Updates(updates).Error
}

func (r *Repository) validateSpeedRequestSubmission(speedRequestID uint, userID uint) error {
	var speedRequest ds.SpeedRequest
	if err := r.db.First(&speedRequest, speedRequestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("заявка не найдена")
		}
		return err
	}

	if speedRequest.CreatorID != userID {
		var user ds.User
		if err := r.db.First(&user, userID).Error; err != nil {
			return errors.New("пользователь не найден")
		}

		if !user.IsModerator {
			return errors.New("403 доступ запрещен") // 403
		}
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

	var moderator ds.User
	if err := r.db.First(&moderator, moderatorID).Error; err != nil {
		return errors.New("модератор не найден")
	}

	if !moderator.IsModerator {
		return errors.New("403 доступ запрещен") // 403
	}

	if speedRequest.Status != ds.StatusSubmitted {
		return errors.New("можно завершать/отклонять только сформированные заявки")
	}

	if status != ds.StatusCompleted && status != ds.StatusRejected {
		return errors.New("неверный статус")
	}

	if status == ds.StatusCompleted {
		if err := r.calculateShipSpeedRequest(speedRequest); err != nil {
			return fmt.Errorf("ошибка расчета скорости контейнеровоза: %w", err)
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
