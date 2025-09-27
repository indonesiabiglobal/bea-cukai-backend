package database

import (
	"Dashboard-TRDP/helper"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error) {
	host := helper.GetEnv("DB_HOST")
	port := helper.GetEnv("DB_PORT")
	username := helper.GetEnv("DB_USERNAME")
	password := helper.GetEnv("DB_PASSWORD")
	dbName := helper.GetEnv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, username, password, dbName, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	return db, err
}
