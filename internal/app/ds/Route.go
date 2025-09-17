package ds

type Route struct {
	RouteID     int    `gorm:"primaryKey"`
	Title       string `gorm:"type:varchar(128)"`
	Distance    int
	Description string `gorm:"type:text;default:null"`
	Delay       int
	ImageUrl    string `gorm:"type:varchar(255);not null"`
	Status      string
}
