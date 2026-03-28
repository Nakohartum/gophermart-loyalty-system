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
	appCtx := context.Background()
	db := db.NewPgDatabase(ConfigData.SecretKey)
	err := db.OpenConnection(appCtx, ConfigData.DatabaseUri)
	if err != nil {
		log.Fatal(err.Error())
	}
	repo := repository.NewDatabaseRepository(db)
	databaseService := service.NewDatabaseRepo(repo)
	authService := service.NewAuthService([]byte(ConfigData.SecretKey), 24 * time.Hour)
	accrualService := service.NewAccrualService(ConfigData.AccrualSystemAddress)
	orderProcessor := service.NewOrderProcessor(databaseService, accrualService, 2*time.Second)
	uh := handler.NewUserHandler(databaseService, authService)
	oh := handler.NewOrderHandler(databaseService, authService)
	mux := http.NewServeMux()
	hanlder := handler.GzipMiddleware(mux)
	authMiddleware := handler.AuthMiddleware(authService)
	mux.Handle("/api/user/register", uh.RegisterUser())
	mux.Handle("/api/user/login", uh.AuthenticateUser())
	mux.Handle("/api/user/orders", authMiddleware(oh.CreateOrder()))

	if ConfigData.AccrualSystemAddress != "" {
		go orderProcessor.Start(appCtx)
	}

	fmt.Println("Server started")
	if err := http.ListenAndServe(ConfigData.RunAddress, hanlder); err != nil {
		fmt.Println(err)
	}
	
}
