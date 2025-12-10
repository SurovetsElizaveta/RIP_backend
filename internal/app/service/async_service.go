package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"rip/internal/app/config"

	"github.com/sirupsen/logrus"
)

type AsyncService struct {
	config config.AsyncServiceConfig
	client *http.Client
}

func NewAsyncService(cfg config.AsyncServiceConfig) *AsyncService {
	return &AsyncService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type CalculateSpeedRequest struct {
	SpeedRequestID uint    `json:"speed_request_id"`
	RouteID        uint    `json:"route_id"`
	Distance       float64 `json:"distance"`
	Delay          float64 `json:"delay"`
	AuthToken      string  `json:"auth_token"`
	DepartureDate  string  `json:"departure_date,omitempty"`
	ArrivalDate    string  `json:"arrival_date,omitempty"`
}

func (s *AsyncService) RequestSpeedCalculation(speedRequestID, routeID uint, distance, delay float64, departureDate, arrivalDate time.Time) error {
	// Форматируем даты в формат "%d.%m.%Y" для Django сервиса
	var departureDateStr, arrivalDateStr string

	if !departureDate.IsZero() {
		departureDateStr = departureDate.Format("02.01.2006")
	}

	if !arrivalDate.IsZero() {
		arrivalDateStr = arrivalDate.Format("02.01.2006")
	}

	payload := CalculateSpeedRequest{
		SpeedRequestID: speedRequestID,
		RouteID:        routeID,
		Distance:       distance,
		Delay:          delay,
		AuthToken:      s.config.Token,
		DepartureDate:  departureDateStr,
		ArrivalDate:    arrivalDateStr,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	req, err := http.NewRequest("POST", s.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.Errorf("Ошибка создания HTTP запроса: %v", err)
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		logrus.Errorf("Ошибка отправки HTTP запроса: %v", err)
		return fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Ошибка чтения тела ответа: %v", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("неожиданный статус ответа: %d, тело: %s", resp.StatusCode, string(body))
	}

	logrus.Infof("Запрос на расчет скорости успешно отправлен: speed_request_id=%d, route_id=%d, departure_date=%s, arrival_date=%s",
		speedRequestID, routeID, departureDateStr, arrivalDateStr)
	return nil
}
