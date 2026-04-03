package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/handler"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/repository/postgres"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/service"
)

func main() {
	parseFlags()
	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("starting gophermart")
	db := postgres.New(ConfigData.SecretKey)
	log.Printf("opening database connection")
	err := db.OpenConnection(appCtx, ConfigData.DatabaseURI)
	if err != nil {
		log.Fatalf("open database connection: %v", err)
	}
	log.Printf("database connection opened")
	authService := service.NewAuthService([]byte(ConfigData.SecretKey), 24*time.Hour)
	appService := service.NewAppService(db, authService)
	accrualService := service.NewAccrualService(ConfigData.AccrualSystemAddress)
	orderProcessor := service.NewOrderProcessor(appService, accrualService, 2*time.Second)
	uh := handler.NewUserHandler(appService, appService)
	oh := handler.NewOrderHandler(appService)
	mux := http.NewServeMux()
	pathHandler := handler.GzipMiddleware(mux)
	authMiddleware := handler.AuthMiddleware(authService)
	mux.Handle("/api/user/register", uh.RegisterUser())
	mux.Handle("/api/user/login", uh.AuthenticateUser())
	mux.Handle("/api/user/orders", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			oh.CreateOrder().ServeHTTP(w, r)
		case http.MethodGet:
			oh.GetListOfUploadedOrders().ServeHTTP(w, r)
		default:
			http.Error(w, "not correct method", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/user/balance", authMiddleware(uh.GetCurrentUserBalance()))
	mux.Handle("/api/user/balance/withdraw", authMiddleware(uh.WithdrawBalance()))
	mux.Handle("/api/user/withdrawals", authMiddleware(uh.GetWithdrawalsInfo()))

	var wg sync.WaitGroup
	if ConfigData.AccrualSystemAddress != "" {
		log.Printf("starting order processor with accrual address %q", ConfigData.AccrualSystemAddress)
		wg.Add(1)
		go func() {
			defer wg.Done()
			orderProcessor.Start(appCtx)
		}()
	}

	srv := &http.Server{
		Addr:    ConfigData.RunAddress,
		Handler: pathHandler,
	}

	log.Printf("http server listening on %q", ConfigData.RunAddress)
	fmt.Println("Server started")
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			log.Printf("http server stopped with error: %v", err)
			fmt.Println(err)
		}
	case <-appCtx.Done():
		log.Printf("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}

	wg.Wait()

	if err := db.CloseConnection(shutdownCtx); err != nil {
		log.Printf("database close error: %v", err)
	}
}
