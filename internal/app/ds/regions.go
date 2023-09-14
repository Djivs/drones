package ds

type Region struct {
	ID         uint `gorm:primaryKey`
	Name       string
	Details    string
	AreaKm     float64
	Population int
}
