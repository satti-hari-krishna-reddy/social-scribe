package v1

import (
	"net/http"
	"time"

	"social-scribe/backend/internal/handlers"
	"social-scribe/backend/internal/middlewares"

	"github.com/gorilla/mux"
)

func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Unprotected routes with IP based rate limiting
	apiV1.Handle("/user/signup",
		middlewares.IPRateLimitMiddleware(10, time.Minute)(http.HandlerFunc(handlers.SignupUserHandler)),
	).Methods(http.MethodPost, http.MethodOptions)
	apiV1.Handle("/user/login",
		middlewares.IPRateLimitMiddleware(10, time.Minute)(http.HandlerFunc(handlers.LoginUserHandler)),
	).Methods(http.MethodPost, http.MethodOptions)
	apiV1.Handle("/user/linkedin-callback",
		middlewares.IPRateLimitMiddleware(30, time.Minute)(http.HandlerFunc(handlers.LinkedCallbackHandler)),
	).Methods(http.MethodGet, http.MethodOptions)
	apiV1.Handle("/user/forgot-password",
		middlewares.IPRateLimitMiddleware(5, time.Minute)(http.HandlerFunc(handlers.ForgotPasswordHandler)),
	).Methods(http.MethodPost, http.MethodOptions)
	apiV1.Handle("/user/reset-password",
		middlewares.IPRateLimitMiddleware(5, time.Minute)(http.HandlerFunc(handlers.ResetPasswordHandler)),
	).Methods(http.MethodPost, http.MethodOptions)

	// Protected routes with user based rate limiting
	apiV1.Handle("/user/blogs",
		middlewares.AuthMiddleware(200, time.Minute, http.HandlerFunc(handlers.GetUserBlogsHandler)),
	).Methods(http.MethodGet, http.MethodOptions)

	apiV1.Handle("/user/logout",
		middlewares.AuthMiddleware(15, time.Minute, http.HandlerFunc(handlers.LogoutUserHandler)),
	).Methods(http.MethodPost, http.MethodOptions)

	apiV1.Handle("/user/notifications",
		middlewares.AuthMiddleware(150, time.Minute, http.HandlerFunc(handlers.GetUserNotificationsHandler)),
	).Methods(http.MethodGet, http.MethodOptions)

	apiV1.Handle("/user/notifications/clear",
		middlewares.AuthMiddleware(20, time.Minute, http.HandlerFunc(handlers.ClearUserNotificationsHandler)),
	).Methods(http.MethodDelete, http.MethodOptions)

	apiV1.Handle("/blogs/schedule",
		middlewares.AuthMiddleware(6, time.Minute, http.HandlerFunc(handlers.ScheduleBlogHandler)),
	).Methods(http.MethodPost, http.MethodOptions)

	apiV1.Handle("/blogs/user/share",
		middlewares.AuthMiddleware(50, time.Minute, http.HandlerFunc(handlers.ShareBlogHandler)),
	).Methods(http.MethodPost, http.MethodOptions)

	apiV1.Handle("/user/scheduled-blogs/cancel",
		middlewares.AuthMiddleware(40, time.Minute, http.HandlerFunc(handlers.CancelScheduledBlogHandler)),
	).Methods(http.MethodDelete, http.MethodOptions)

	apiV1.Handle("/user/connect-twitter",
		middlewares.AuthMiddleware(15, time.Minute, http.HandlerFunc(handlers.ConnectXhandler)),
	).Methods(http.MethodGet, http.MethodOptions)

	apiV1.Handle("/user/twitter-callback",
		middlewares.AuthMiddleware(10, time.Minute, http.HandlerFunc(handlers.XcallbackHandler)),
	).Methods(http.MethodGet, http.MethodOptions)

	apiV1.Handle("/user/connect-linkedin",
		middlewares.AuthMiddleware(15, time.Minute, http.HandlerFunc(handlers.ConnectLinkedInHandler)),
	).Methods(http.MethodGet, http.MethodOptions)

	apiV1.Handle("/user/getinfo", middlewares.AuthMiddleware(15, time.Minute, http.HandlerFunc(handlers.GetUserInfoHandler))).Methods(http.MethodGet, http.MethodOptions)

	apiV1.Handle("/user/verify-hashnode",
		middlewares.AuthMiddleware(10, time.Minute, http.HandlerFunc(handlers.VerifyHashnodeHandler)),
	).Methods(http.MethodPost, http.MethodOptions)

	apiV1.Handle("/user/verify-email",
		middlewares.AuthMiddleware(10, time.Minute, http.HandlerFunc(handlers.VerifyEmailHandler)),
	).Methods(http.MethodPost, http.MethodOptions)

	apiV1.Handle("/user/resend-otp",
		middlewares.AuthMiddleware(5, time.Minute, http.HandlerFunc(handlers.ResetEmailOtpHandler)),
	).Methods(http.MethodGet, http.MethodOptions)
	apiV1.Handle("/user/delete-account",
		middlewares.AuthMiddleware(20, time.Minute, http.HandlerFunc(handlers.DeleteAccountHandler)),
	).Methods(http.MethodDelete, http.MethodOptions)

	return router
}
