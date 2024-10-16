package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"social-scribe/backend/api/v1"
	repo "social-scribe/backend/internal/repositories"
)

func main() {
	repo.InitMongoDb()
    router := v1.RegisterRoutes()
    hostname, err := os.Hostname()
	if err != nil {
		hostname = "MISSING"
	}

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		log.Printf("[DEBUG] Running on %s:9696", hostname)
		log.Fatal(http.ListenAndServe(":9696", router))
	} else {
		log.Printf("[DEBUG] Running on %s:%s", hostname, port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
	}
}
