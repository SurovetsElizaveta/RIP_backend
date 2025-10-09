package ds

type Route struct {
	RouteID     uint   `gorm:"primaryKey;autoIncrement"`
	Title       string `gorm:"size:128;not null"`
	Distance    int    `gorm:"not null"`
	Description string `gorm:"type:text;not null"`
	ImageURL    string `gorm:"size:255;default:null"`
	Delay       int    `gorm:"not null"`
	Status      string `gorm:"size:20;not null"`
}
