package database

import (
	"github.com/victor-butita/savannah-challenge/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Customer{}, &models.Order{})
	if err != nil {
		return nil, err
	}

	return db, nil
}