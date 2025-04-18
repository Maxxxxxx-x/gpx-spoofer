package main

import (
	"log"
	"os"

	open_elevation "github.com/Maxxxxxx-x/gpx-spoofer/api/OE"
	open_route_service "github.com/Maxxxxxx-x/gpx-spoofer/api/ORS"
	"github.com/Maxxxxxx-x/gpx-spoofer/config"
	"github.com/Maxxxxxx-x/gpx-spoofer/db"
	"github.com/Maxxxxxx-x/gpx-spoofer/spoofing"
)


func main() {
	log.Println("Starting spoofer")
	config, err := config.GetFromEnv(".env")
	if err != nil {
		log.Fatalf("Error occured while reading env file: %s", err.Error())
		os.Exit(1)
	}

	routeApi := open_route_service.New(config.RouteAPI)
	elevApi := open_elevation.New(config.ElevApi)

	dbConn, err := db.CreateDatabaseConnection(config.Database)
	if err != nil {
		log.Fatalf("Error occured while creating db connection: %s", err.Error())
		os.Exit(1)
	}
	defer dbConn.Close()
	log.Println("Connected to database!")

	db := db.New(dbConn)


	spoofer := spoofing.New(routeApi, elevApi, db)
	spoofer.Start()
	log.Println("Completed")
}
