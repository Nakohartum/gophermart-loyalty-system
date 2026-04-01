package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/config/db"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"github.com/Nakohartum/gophermart-loyalty-system/internal/service"
)

type UserHandler struct {
	databaseService *service.DatabaseService
	authService     *service.AuthService
}

func NewUserHandler(dbService *service.DatabaseService, authService *service.AuthService) *UserHandler {
	return &UserHandler{
		databaseService: dbService,
		authService: authService,
	}
}

func (uh *UserHandler) RegisterUser() http.HandlerFunc {
	fun := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "not correct method", http.StatusBadRequest)
			return
		}
		var req model.AuthRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password are required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		userId, err := uh.databaseService.RegisterUser(ctx, req.Login, req.Password)

		if err != nil {
			if errors.Is(err, db.ErrUserAlreadyExists) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := uh.authService.GenerateToken(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
		w.Header().Set("Authorization", "Bearer " + token)

		w.WriteHeader(http.StatusOK)
	}
	return fun
}

func (uh *UserHandler) AuthenticateUser() http.HandlerFunc {
	fun := func (w http.ResponseWriter, r *http.Request)  {
		if (r.Method != http.MethodPost) {
			http.Error(w, "not correct method", http.StatusBadRequest)
			return 
		}

		var req model.AuthRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password are required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		userId, err := uh.databaseService.AuthenticateUser(ctx, req.Login, req.Password)

		if err != nil {
			if errors.Is(err, db.ErrPasswordNotMatch){
				http.Error(w, "passwords not match", http.StatusUnauthorized)
				return 
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}

		token, err := uh.authService.GenerateToken(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
		w.Header().Set("Authorization", "Bearer " + token)

		w.WriteHeader(http.StatusOK)
	}

	return fun
}

func (uh *UserHandler) GetCurrentUserBalance() http.HandlerFunc {
	fun := func (w http.ResponseWriter, r *http.Request)  {
		if r.Method != http.MethodGet {
			http.Error(w, "not correct method", http.StatusBadRequest)
			return 
		}
		userId, ok := UserIDFromContext(r.Context())

		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return 
		}
		
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		userBalance := uh.databaseService.GetCurrentUserBalance(ctx, userId)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userBalance); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return 
		}
		w.WriteHeader(http.StatusOK)
	}
	return fun 
}
