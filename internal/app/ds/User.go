package ds

type User struct {
	UserID      uint   `gorm:"primaryKey;autoIncrement"`
	Login       string `gorm:"unique;size:128;not null"`
	Password    string `gorm:"size:128;not null"`
	IsModerator bool   `gorm:"not null"`
}
