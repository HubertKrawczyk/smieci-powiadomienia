package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"smieci-sms/config"
	"smieci-sms/internal/api"
	"smieci-sms/internal/repository"
	"smieci-sms/internal/scheduler"
	"smieci-sms/internal/service"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

func main() {
	cfg := config.LoadConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("CRITICAL CONFIG ERROR: %v. Server shutting down.", err)
	}

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := initDatabase(db); err != nil {
		log.Fatalf("failed to initialize database schema: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	// smsService := service.NewSMSService(cfg.SMSProviderAPIKey)
	garbageService := service.NewGarbageService(cfg.CityGarbageURL)
	telegramSvc := service.NewTelegramService(cfg.TelegramBotToken)
	appScheduler := scheduler.NewScheduler(userRepo, garbageService, nil, telegramSvc)

	router := api.NewRouter(userRepo, garbageService,
		cfg.TelegramSecretToken, telegramSvc, cfg.AdminSecret)
	appScheduler.ScheduleDailyTasks()

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	go func() {
		log.Printf("starting server on %s", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}

func initDatabase(db *repository.DB) error {
	entries, err := sqlFiles.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("failed to read embedded sql directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := "sql/" + entry.Name()
		log.Printf("Executing SQL script: %s", filePath)

		content, err := sqlFiles.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		if _, err := db.Conn.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute sql script %s: %w", filePath, err)
		}
	}
	return nil
}
