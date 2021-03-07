package database

import (
	"os"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// The main database instance.
var Instance *pg.DB

// A default database url that is used to bind the postgres service.
const defaultDatabaseURL = "postgres://postgres:password@localhost:5432/peerbridge?sslmode=disable"

// Get a database url from the process environment variables.
// This method is used as a part of database initialization.
// The database url can be configured by setting the
// environment variable `DATABASE_URL`.
func getDatabaseURL() string {
	port := os.Getenv("DATABASE_URL")
	if port != "" {
		return port
	}

	return defaultDatabaseURL
}

// Initialize the database and create tables for given models.
func Initialize(models []interface{}) error {
	opt, err := pg.ParseURL(getDatabaseURL())
	if err != nil {
		return err
	}

	Instance = pg.Connect(opt)

	for _, model := range models {
		err := Instance.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
