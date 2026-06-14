package api

import (
	"net/http"

	"smieci-sms/internal/repository"
	"smieci-sms/internal/service"

	"github.com/gorilla/mux"
)

func NewRouter(repo repository.UserRepository, garbageSvc service.GarbageService, telegramSecretToken string, telegramSvc service.TelegramService, adminSecret string) *mux.Router {
	router := mux.NewRouter()
	handler := NewHandler(repo, garbageSvc)
	telegramHandler := NewTelegramHandler(repo, garbageSvc, telegramSecretToken, telegramSvc)

	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token != "Bearer "+adminSecret {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}

	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/users", authMiddleware(handler.ListUsers)).Methods("GET")
	router.HandleFunc("/schedules/fetch", authMiddleware(handler.FetchSchedules)).Methods("POST")
	router.HandleFunc("/telegram/start", telegramHandler.Start).Methods("POST")

	return router
}
