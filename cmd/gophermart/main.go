package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/config/db"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/handler"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/repository"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/service"
)

func main() {
	parseFlags()
	db := db.NewPgDatabase(ConfigData.SecretKey)
	err := db.OpenConnection(context.Background(), ConfigData.DatabaseUri)
	if err != nil {
		log.Fatal(err.Error())
	}
	repo := repository.NewDatabaseRepository(db)
	databaseService := service.NewDatabaseRepo(repo)
	authService := service.NewAuthService([]byte(ConfigData.SecretKey), 24 * time.Hour)
	uh := handler.NewUserHandler(databaseService, authService)
	mux := http.NewServeMux()
	hanlder := handler.GzipMiddleware(mux)
	mux.Handle("/user/register", uh.RegisterUser())
	mux.Handle("/user/login", uh.AuthenticateUser())
	fmt.Println("Server started")
	if err := http.ListenAndServe(ConfigData.RunAddress, hanlder); err != nil {
		fmt.Println(err)
	}
	
}
