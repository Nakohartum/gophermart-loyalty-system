package db

import "context"

type Storage interface {
	OpenConnection(ctx context.Context, databaseUri string) error
	CloseConnection(ctx context.Context) error
	CheckConnection(ctx context.Context) error
	
}