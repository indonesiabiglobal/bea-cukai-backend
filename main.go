package main

import (
	"Dashboard-TRDP/database"
	"Dashboard-TRDP/helper"
	"Dashboard-TRDP/routes"
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
