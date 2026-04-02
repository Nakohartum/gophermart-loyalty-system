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
	log.Printf("starting gophermart")
	db := db.NewPgDatabase(ConfigData.SecretKey)
	log.Printf("opening database connection")
	err := db.OpenConnection(appCtx, ConfigData.DatabaseUri)
	if err != nil {
		log.Fatalf("open database connection: %v", err)
	}
	log.Printf("database connection opened")
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
	mux.Handle("/api/user/orders", authMiddleware(oh.Orders()))
	mux.Handle("/api/user/balance", authMiddleware(uh.GetCurrentUserBalance()))
	mux.Handle("/api/user/balance/withdraw", authMiddleware(uh.WithdrawBalance()))
	mux.Handle("/api/user/withdrawals", authMiddleware(uh.GetWithdrawalsInfo()))
	if ConfigData.AccrualSystemAddress != "" {
		log.Printf("starting order processor with accrual address %q", ConfigData.AccrualSystemAddress)
		go orderProcessor.Start(appCtx)
	}

	log.Printf("http server listening on %q", ConfigData.RunAddress)
	fmt.Println("Server started")
	if err := http.ListenAndServe(ConfigData.RunAddress, hanlder); err != nil {
		log.Printf("http server stopped: %v", err)
		fmt.Println(err)
	}
	
}
