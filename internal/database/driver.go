package database

import (
	"gorm.io/gorm"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
)

type DriverFunc func (dsn string) gorm.Dialector

// todo: assign drivers correctly once packages are installed
type DriverStructure struct {
	Sqlite DriverFunc
	Postgres DriverFunc
}

var Drivers = DriverStructure{gormlite.Open, nil}