package ds

import "time"

type RequestRoute struct {
	RequestID   int `gorm:"primaryKey;not null"`
	RouteID     int `gorm:"primaryKey;not null"`
	ArrivalDate time.Time
	ShipSpeed   int

	Request Request `gorm:"foreignKey:RequestID"`
	Route   Route   `gorm:"foreignKey:RouteID"`
}
