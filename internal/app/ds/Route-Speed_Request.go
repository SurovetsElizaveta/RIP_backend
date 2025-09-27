package ds

import "time"

type RouteSpeedRequest struct {
	SpeedRequestID int `gorm:"primaryKey;not null"`
	RouteID        int `gorm:"primaryKey;not null"`
	ArrivalDate    time.Time
	ShipSpeed      int

	Request SpeedRequest `gorm:"foreignKey:SpeedRequestID"`
	Route   Route        `gorm:"foreignKey:RouteID"`
}
