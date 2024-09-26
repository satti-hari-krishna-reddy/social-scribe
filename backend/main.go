package main

import (
    "log"
    "net/http"
    "social-scribe/backend/api/v1" 
)

func main() {
    router := v1.RegisterRoutes()
    log.Fatal(http.ListenAndServe(":8080", router))
}
