package ds

type Route struct {
	RouteID     uint   `gorm:"primaryKey;autoIncrement"`
	Title       string `gorm:"size:128"`
	Distance    int
	Description string `gorm:"type:text"`
	ImageURL    string `gorm:"size:255"`
	Delay       int
	Status      string `gorm:"size:20"`
}
