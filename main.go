package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"single-win-system/configs"
	"single-win-system/handlers"

	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	logrus "github.com/sirupsen/logrus"
)

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	githubUsername, exists := os.LookupEnv("POSTGRES_DSN_HOST")

	if exists {
		fmt.Println(githubUsername)
	}
	configs.Db = pg.Connect(&pg.Options{
		Addr:     os.Getenv("POSTGRES_DSN_HOST"),
		User:     os.Getenv("POSTGRES_DSN_USER"),
		Password: os.Getenv("POSTGRES_DSN_PASSWORD"),
		Database: os.Getenv("POSTGRES_DSN_DB"),
	})
	ctx := context.Background()
	if err := configs.Db.Ping(ctx); err != nil {
		logrus.Info("DB connect error")
	} else {
		logrus.Info("DB is connect")
	}

	logrus.Info("Starting the motiv-single-win-system service...")

	router := handlers.Router()
	logrus.Info("The motiv-single-win-system service is ready to listen and serve.")
	go log.Print(http.ListenAndServe(":3000", router))
}
