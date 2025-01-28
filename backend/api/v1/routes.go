package v1

import (
	"net/http"
	"social-scribe/backend/internal/handlers"

	"github.com/gorilla/mux"
)

func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	apiV1.HandleFunc("/user/signup", handlers.SignupUserHandler).Methods(http.MethodPost)
	apiV1.HandleFunc("/user/login", handlers.LoginUserHandler).Methods(http.MethodPost)
	apiV1.HandleFunc("/user/scheduled_posts", handlers.GetUserScheduledBlogsHandler).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/user/blogs", handlers.GetUserBlogsHandler).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/user/getinfo", handlers.GetUserInfoHandler).Methods(http.MethodGet)

	apiV1.HandleFunc("/user/notifications", handlers.GetUserNotificationsHandler).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/user/notifications/clear", handlers.ClearUserNotificationsHandler).Methods(http.MethodDelete, http.MethodOptions)

	apiV1.HandleFunc("/blogs/schedule", handlers.ScheduleUserBlogHandler).Methods(http.MethodPost, http.MethodOptions)
	apiV1.HandleFunc("/blogs/schedule/delete", handlers.GetUserSharedBlogsHandler).Methods(http.MethodDelete, http.MethodOptions)
	apiV1.HandleFunc("/blogs/user/share", handlers.ShareBlogHandler).Methods(http.MethodPost, http.MethodOptions)
	apiV1.HandleFunc("/blogs/user/shared-blogs", handlers.GetUserSharedBlogsHandler).Methods(http.MethodGet, http.MethodOptions)

	apiV1.HandleFunc("/user/connect-twitter", handlers.ConnectXhandler).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/user/twitter-callback", handlers.XcallbackHandler).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/user/connect-linkedin", handlers.ConnectLinkedInHandler).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/user/linkedin-callback", handlers.LinkedCallbackHandler).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/user/verify-hashnode", handlers.VerifyHashnodeHandler).Methods(http.MethodPost, http.MethodOptions)

	return router

}
