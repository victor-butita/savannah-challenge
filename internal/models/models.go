package models

import "time"

type Customer struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Code        string `gorm:"not null;unique"`
	PhoneNumber string `gorm:"not null"`
	CreatedAt   time.Time
}

type Order struct {
	ID         uint `gorm:"primaryKey"`
	Item       string
	Amount     float64
	Time       time.Time
	CustomerID uint
	Customer   Customer `gorm:"foreignKey:CustomerID"`
	CreatedAt  time.Time
}
