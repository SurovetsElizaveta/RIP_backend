package ds

import "time"

type RouteSpeedRequest struct {
	RouteID        uint      `gorm:"primaryKey"`
	SpeedRequestID uint      `gorm:"primaryKey"`
	ArrivalDate    time.Time `gorm:"default:null"`
	ShipSpeed      int       `gorm:"default:null"`

	Route        Route        `gorm:"foreignKey:RouteID"`
	SpeedRequest SpeedRequest `gorm:"foreignKey:SpeedRequestID"`
}
