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
	SpeedRequestID uint      `gorm:"primaryKey;autoIncrement"`
	DepartureDate  time.Time `gorm:"default:null"`
	CreationDate   time.Time `gorm:"not null"`
	FormationDate  time.Time `gorm:"default:null"`
	CompletionDate time.Time `gorm:"default:null"`
	CreatorID      uint      `gorm:"not null"`
	ModeratorID    uint      `gorm:"default:null"`
	Status         string    `gorm:"type:varchar(20);not null"`

	Creator   User `gorm:"foreignKey:CreatorID;references:UserID"`
	Moderator User `gorm:"foreignKey:ModeratorID;references:UserID"`
}
