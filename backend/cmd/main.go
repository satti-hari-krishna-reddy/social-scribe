package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"social-scribe/backend/api/v1"
	"social-scribe/backend/internal/handlers"
	repo "social-scribe/backend/internal/repositories"
	"social-scribe/backend/internal/scheduler"

	"github.com/rs/cors"
)

func setupCors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://192.168.29.3:9696", "http://192.168.29.3:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	})
}

func main() {
	repo.InitMongoDb()
	router := v1.RegisterRoutes()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "MISSING"
	}
	taskScheduler := scheduler.NewScheduler()
	handlers.InitScheduler(taskScheduler)
	corsHandler := setupCors()

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		log.Printf("[DEBUG] Running on %s:9696", hostname)
		log.Fatal(http.ListenAndServe(":9696", corsHandler.Handler(router)))
	} else {
		log.Printf("[DEBUG] Running on %s:%s", hostname, port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), corsHandler.Handler(router)))
	}
}
