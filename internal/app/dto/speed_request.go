package dto

import (
	"fmt"
	"rip/internal/app/ds"
	"time"
)

type SpeedRequest struct {
	SpeedRequestID uint      `json:"id"`
	DepartureDate  time.Time `json:"departure_date"`
	CreationDate   time.Time `json:"creation_date"`
	FormationDate  time.Time `json:"formation_date"`
	CompletionDate time.Time `json:"completion_date"`
	CreatorLogin   string    `json:"creator_login"`
	ModeratorLogin string    `json:"moderator_login,omitempty"`
	Status         string    `json:"status"`
}

type SpeedRequestResponse struct {
	SpeedRequestID uint      `json:"id"`
	DepartureDate  time.Time `json:"departure_date"`
	CreationDate   time.Time `json:"creation_date"`
	FormationDate  time.Time `json:"formation_date"`
	CompletionDate time.Time `json:"completion_date"`
	Status         string    `json:"status"`
}

type SpeedRequestDetailedResponse struct {
	SpeedRequest SpeedRequest            `json:"speed_request"`
	RouteReq     []RouteSpeedRequestInfo `json:"route_req"`
	Routes       []RouteInfo             `json:"routes"`
	Result       int                     `json:"result"`
}

type RouteSpeedRequestInfo struct {
	ArrivalDate time.Time `json:"arrival_date"`
	ShipSpeed   int       `json:"ship_speed"`
}

type RouteInfo struct {
	RouteID     uint   `json:"route_id"`
	Title       string `json:"title"`
	Distance    int    `json:"distance"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Status      string `json:"status"`
}

type UpdateSpeedRequest struct {
	DepartureDate *time.Time `json:"departure_date" binding:"omitempty"`
}

type SpeedRequestFilter struct {
	Status   *string    `form:"status"`
	DateFrom *time.Time `form:"date_from"`
	DateTo   *time.Time `form:"date_to"`
}

func (f *SpeedRequestFilter) Validate() error {
	if f.DateFrom != nil && f.DateTo != nil && f.DateFrom.After(*f.DateTo) {
		return fmt.Errorf("date_from не может быть после date_to")
	}

	if f.Status != nil {
		validStatuses := map[string]bool{
			ds.StatusSubmitted: true,
			ds.StatusCompleted: true,
			ds.StatusRejected:  true,
		}
		if !validStatuses[*f.Status] {
			return fmt.Errorf("неверный статус")
		}
	}

	return nil
}
