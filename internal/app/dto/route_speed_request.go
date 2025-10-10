package dto

type RouteSpeedRequest struct {
	RouteID        uint   `json:"route_id"`
	SpeedRequestID uint   `json:"speed_request_id"`
	ArrivalDate    string `json:"arrival_date"`
	ShipSpeed      int    `json:"ship_speed"`
}

type RemoveRouteSpeedRequest struct {
	SpeedRequestID int `json:"speed_request_id" binding:"required"`
	RouteID        int `json:"route_id" binding:"required"`
}

type UpdateRouteSpeedRequest struct {
	SpeedRequestID int    `json:"speed_request_id"`
	RouteID        int    `json:"route_id"`
	ArrivalDate    string `json:"arrival_date,omitempty"`
}
