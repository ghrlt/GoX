package main

import (
	"fmt"
	"os"

	"gox/database"
	server "gox/routes"
	"gox/utils"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Starting Masskaa App...")
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ~ Charge les variables d’environnement depuis .env.dev(.local) ou .env.prod
	if err := godotenv.Load(".env"); err != nil {
		utils.ConsoleLog("❌ Error loading .env file: %v", err).Fatal()
	}

	if utils.GetEnv("GO_ENV", "dev") == "dev" {
		utils.ConsoleLog("🚧 Loading DEV environment variables")

		if err := godotenv.Load(".env.dev", ".env.dev.local"); err != nil {
			utils.ConsoleLog("❌ Error loading .env.dev(.local) file: %v", err).Fatal()
		}
	} else {
		utils.ConsoleLog("🚀 Loading PROD environment variables")

		if err := godotenv.Load(".env.prod"); err != nil {
			utils.ConsoleLog("❌ Error loading .env.prod file: %v", err).Fatal()
		}
	}

	// Récupère les variables d’environnement
	dbHost := utils.GetEnv("POSTGRES_HOST", "localhost")
	dbPort := utils.GetEnv("POSTGRES_PORT", "5432")
	dbUser := utils.GetEnv("POSTGRES_USER", "postgres")
	dbPassword := utils.GetEnv("POSTGRES_PASSWORD", "password")
	dbName := utils.GetEnv("POSTGRES_DB", "gox")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)
	database.InitDB(dsn)
	server.Start()
	return nil
}
