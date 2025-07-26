package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection() (*gorm.DB, error) {
	dsn := "your_connection_string_here"
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
