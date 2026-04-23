package model

type Driver struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255"`
	Status    int       `gorm:"default:0"`
}

func (Driver) TableName() string { return "" }
