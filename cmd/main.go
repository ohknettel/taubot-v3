package main

import (
	"log"
	"os"
	"os/signal"
	
	"github.com/joho/godotenv"
	"github.com/ohknettel/taubot-v3/internal/bot"
	"github.com/ohknettel/taubot-v3/internal/database"
)

var bot_logger = log.New(os.Stderr, "[BOT] ", log.Ldate|log.Ltime)
var db_logger = log.New(os.Stderr, "[BACKEND] ", log.Ldate|log.Ltime)
var main_logger = log.New(os.Stderr, "[MAIN] ", log.Ldate|log.Ltime)

func main() {
	err := godotenv.Load()
	if err != nil {
		main_logger.Fatalf("An error occured while loading environmental variables: %v", err)
		return
	}

	token := os.Getenv("token")
	if token == "" {
		main_logger.Fatal("Discord bot token variable 'token' not found. Please provide token=... as an environmental variable.")
		return
	}

	database_uri := os.Getenv("database_uri")
	if database_uri == "" {
		database_uri = "database.db"
	}

	driver := os.Getenv("database_driver")
	var db_driver database.DriverFunc
	switch driver {
	case "sqlite":
		db_driver = database.Drivers.Sqlite
	case "postgres":
		db_driver = database.Drivers.Postgres
	default:
		db_driver = database.Drivers.Sqlite
	}

	bot, err := bot.NewBot(token)
	if err != nil {
		main_logger.Fatalf("An error occured while creating a bot instance: %v", err)
		return
	}

	bot.SetLogger(bot_logger)

	err = bot.RegisterHandlers()
	if err != nil {
		main_logger.Fatalf("An error occured while registering handlers: %v", err)
		return
	}

	err = bot.Run()
	if err != nil {
		main_logger.Fatalf("An error occured while opening the session: %v", err)
	}
	
	err = bot.SetupBackend(database_uri, db_driver, db_logger)

	if err != nil {
		main_logger.Fatalf("An error occured while setting up the backend: %v", err)
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	main_logger.Printf("Shutting down...")
}