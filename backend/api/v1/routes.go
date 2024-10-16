package v1

import (
    "social-scribe/backend/internal/handlers"
	"github.com/gorilla/mux"
)

func RegisterRoutes() *mux.Router {
    router := mux.NewRouter()
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	apiV1.HandleFunc("/user/signup", handlers.SignupUserHandler).Methods("POST")
	apiV1.HandleFunc("/user/login", handlers.LoginUserHandler).Methods("POST")
	apiV1.HandleFunc("/user/{id}/shared_posts", handlers.GetUserSharedBlogsHandler).Methods("GET")
	apiV1.HandleFunc("/user/{id}/scheduled_posts", handlers.GetUserScheduledBlogsHandler).Methods("GET")
	apiV1.HandleFunc("/user/{id}/notifications", handlers.GetUserNotificationsHandler).Methods("GET")
	apiV1.HandleFunc("/user/{id}/notifications/clear", handlers.ClearUserNotificationsHandler).Methods("PUT")

	apiV1.HandleFunc("/blogs/share", handlers.GetUserSharedBlogsHandler).Methods("POST")
	apiV1.HandleFunc("/blogs/schedule", handlers.GetUserSharedBlogsHandler).Methods("POST")
	apiV1.HandleFunc("/blogs/schedule/delete", handlers.GetUserSharedBlogsHandler).Methods("DELETE")

	return router

}
