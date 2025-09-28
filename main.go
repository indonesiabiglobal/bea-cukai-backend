package main

import (
	"Bea-Cukai/database"
	"Bea-Cukai/helper"
	"Bea-Cukai/routes"
	"log"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Error connect to database", err)
		panic(err)
	}

	app := routes.NewRoute(db)

	apiPort := helper.GetEnv("PORT")
	log.Fatal(app.Run(":" + apiPort))
}
