package db

import "errors"

var (
	errNoConnectionToClose         = errors.New("no connection to close")
	errNoConnectionToRunMigrations = errors.New("no connection to run migrations")
	errPasswordNotMatch            = errors.New("passwords do not match")
	errOrderNotFound               = errors.New("order not found")
	errUserNotFound                = errors.New("user not found")
	errOrderAlreadyProcessed       = errors.New("order already processed")
)
