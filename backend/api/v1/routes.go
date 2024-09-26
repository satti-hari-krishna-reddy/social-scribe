package v1

import (
    "social-scribe/backend/internal/handlers"
	"github.com/gorilla/mux"
)

func RegisterRoutes() *mux.Router {
    router := mux.NewRouter()
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	apiV1.HandleFunc("/user/register", handlers.RegisterUserHandler).Methods("GET")
	apiV1.HandleFunc("/user/login", handlers.UserLoginHandler).Methods("GET")
	apiV1.HandleFunc("/user", handlers.GetUserHandler).Methods("GET")
	apiV1.HandleFunc("/user/shared_posts", handlers.UserLoginHandler).Methods("GET")
	apiV1.HandleFunc("/user/scheduled_posts", handlers.UserLoginHandler).Methods("GET")

	apiV1.HandleFunc("/notifications", handlers.UserLoginHandler).Methods("GET")
	apiV1.HandleFunc("/notifications/clear", handlers.UserLoginHandler).Methods("GET")

	return router

}

