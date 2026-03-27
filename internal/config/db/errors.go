package db

import "errors"

var (
	ErrNoConnectionToClose         = errors.New("no connection to close")
	ErrNoConnectionToRunMigrations = errors.New("no connection to run migrations")
	ErrPasswordNotMatch            = errors.New("passwords do not match")
	ErrOrderNotFound               = errors.New("order not found")
	ErrUserNotFound                = errors.New("user not found")
	ErrOrderAlreadyProcessed       = errors.New("order already processed")
	ErrUserAlreadyExists           = errors.New("user already exists")
)
