package handlers

import (
	"log"
	"github.com/ohknettel/taubot-v3/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func PrepareDatabase(uri string, driver database.DriverFunc, db_logger *log.Logger, log_level logger.LogLevel) (*gorm.DB, error) {
	models := database.Models
	db, err := gorm.Open(driver(uri), &gorm.Config{
		PrepareStmt: true,
		SkipDefaultTransaction: true,
		Logger: logger.New(
			db_logger,
			logger.Config{
	            LogLevel:                  log_level,
	            IgnoreRecordNotFoundError: true,
	            Colorful:                  false,
			},
		),
	})

	if err == nil {
		db.AutoMigrate(models...)
		return db, err
	} else {
		return  nil, err
	}
}