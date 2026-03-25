package db

import "errors"

var (
	errNoConnectionToClose         = errors.New("No connection to close")
	errNoConnectionToRunMigrations = errors.New("No connection to run migrations")
)
