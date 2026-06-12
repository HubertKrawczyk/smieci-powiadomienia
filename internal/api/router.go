package api

import (
	"smieci-sms/internal/repository"
	"smieci-sms/internal/service"

	"github.com/gorilla/mux"
)

func NewRouter(repo repository.UserRepository, garbageSvc service.GarbageService, telegramSecretToken string, telegramBotToken string) *mux.Router {
	router := mux.NewRouter()
	handler := NewHandler(repo, garbageSvc)
	telegramHandler := NewTelegramHandler(repo, garbageSvc, telegramSecretToken, telegramBotToken)

	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/users", handler.CreateUserLocation).Methods("POST")
	router.HandleFunc("/users", handler.ListUsers).Methods("GET")
	router.HandleFunc("/schedules/fetch", handler.FetchSchedules).Methods("POST")
	router.HandleFunc("/telegram/start", telegramHandler.Start).Methods("POST")

	return router
}
