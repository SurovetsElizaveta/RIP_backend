package ds

type User struct {
	UserID      uint   `gorm:"primaryKey;autoIncrement"`
	Login       string `gorm:"unique;size:128"`
	Password    string `gorm:"size:128"`
	IsModerator bool
}
