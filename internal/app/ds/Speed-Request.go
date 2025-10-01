package ds

import (
	"time"
)

const (
	StatusDraft     = "черновик"
	StatusSubmitted = "сформирована"
	StatusCompleted = "завершена"
	StatusRejected  = "отклонена"
	StatusDeleted   = "удалена"
)

type SpeedRequest struct {
	SpeedRequestID uint `gorm:"primaryKey;autoIncrement"`
	DepartureDate  time.Time
	CreationDate   time.Time `gorm:"not null"`
	FormationDate  time.Time
	CompletionDate time.Time
	CreatorID      uint `gorm:"not null"`
	ModeratorID    uint
	Status         string `gorm:"type:varchar(20);not null"`

	Creator   User `gorm:"foreignKey:CreatorID;references:UserID"`
	Moderator User `gorm:"foreignKey:ModeratorID;references:UserID"`
}
