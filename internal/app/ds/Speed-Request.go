package ds

import (
	"time"
)

type SpeedRequest struct {
	SpeedRequestID int `gorm:"primaryKey"`
	DepartureDate  time.Time
	CreationDate   time.Time `gorm:"not null"`
	FormationDate  time.Time
	CompletionDate time.Time
	Status         string `gorm:"type:varchar(20);not null"`
	CreatorID      int    `gorm:"not null"`
	ModeratorID    int

	Creator   Users `gorm:"foreignKey:CreatorID"`
	Moderator Users `gorm:"foreignKey:ModeratorID"`
}
