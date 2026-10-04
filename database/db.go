package database

import (
	"Bea-Cukai/helper"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error) {
	return connectMySQL("DB_")
}

// ConnectFKI connects to the live FKI database (MySQL).
func ConnectFKI() (*gorm.DB, error) {
	return connectMySQL("FKI_DB_")
}

// ConnectItory connects to the live Itory database (PostgreSQL).
func ConnectItory() (*gorm.DB, error) {
	// prefer = TLS if the server offers it; set verify-full (+ sslrootcert) once the server has a CA cert.
	sslMode := helper.GetEnv("ITORY_DB_SSLMODE")
	if sslMode == "" {
		sslMode = "prefer"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=10 TimeZone=Asia/Jakarta",
		helper.GetEnv("ITORY_DB_HOST"), helper.GetEnv("ITORY_DB_PORT"), helper.GetEnv("ITORY_DB_USERNAME"),
		helper.GetEnv("ITORY_DB_PASSWORD"), helper.GetEnv("ITORY_DB_NAME"), sslMode)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func connectMySQL(prefix string) (*gorm.DB, error) {
	host := helper.GetEnv(prefix + "HOST")
	port := helper.GetEnv(prefix + "PORT")
	username := helper.GetEnv(prefix + "USERNAME")
	password := helper.GetEnv(prefix + "PASSWORD")
	dbName := helper.GetEnv(prefix + "NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
		username, password, host, port, dbName)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
