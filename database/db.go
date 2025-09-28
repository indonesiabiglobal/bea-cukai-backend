package database

import (
	"Bea-Cukai/helper"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error) {
	host := helper.GetEnv("DB_HOST")
	port := helper.GetEnv("DB_PORT")
	username := helper.GetEnv("DB_USERNAME")
	password := helper.GetEnv("DB_PASSWORD")
	dbName := helper.GetEnv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	return db, err
}
