package ds

type Users struct {
	UserID      int    `gorm:"primary_key" json:"user_id"`
	Login       string `gorm:"type:varchar(128);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}
