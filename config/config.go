package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Database struct {
	Host         string
	Port         string
	Login        string
	Password     string
	DatabaseName string
}

type Config struct {
	RouteAPI string
	ElevApi  string
	Database  Database
}

func GetFromEnv(path string) (Config, error) {
	err := godotenv.Load(path)
	if err != nil {
		return Config{}, errors.New("Failed to load ENV file")
	}

	routeApi := os.Getenv("ROUTE_API")
	if routeApi == "" {
		return Config{}, errors.New("ROUTE_API not set!")
	}

	elevApi := os.Getenv("ELEVATE_API")
	if elevApi == "" {
		return Config{}, errors.New("ELEVATE_API not set!")
	}

	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		return Config{}, errors.New("DATABASE_HOST not set!")
	}

	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		return Config{}, errors.New("DATABASE_PORT not set!")
	}

	dbLogin := os.Getenv("DATABASE_LOGIN")
	if dbLogin == "" {
		return Config{}, errors.New("DATABASE_LOGIN not set!")
	}

	dbPass := os.Getenv("DATABASE_PASSWORD")
	if dbPass == "" {
		return Config{}, errors.New("DATABASE_PASSWORD not set!")
	}

	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		return Config{}, errors.New("DATABASE_NAME not set!")
	}

	database := Database{
		Host: dbHost,
		Port: dbPort,
		Login: dbLogin,
		Password: dbPass,
		DatabaseName: dbName,
	}

	config := Config{
		RouteAPI: routeApi,
		ElevApi:  elevApi,
		Database: database,
	}

	return config, nil
}
