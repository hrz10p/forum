package main

import (
	"context"
	"fmt"
	"forum/pkg/models"
	"forum/pkg/services"
	"forum/pkg/utils/cookies"
	"forum/pkg/utils/logger"
	"net/http"
	"time"
)

type contextKey string

var contextKeyUser = contextKey("activeUser")

type Middle struct {
	Service *services.Service
}

func NewMiddle(Service *services.Service) *Middle {
	return &Middle{
		Service: Service,
	}
}

func (app *Middle) Authenticate(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := cookies.GetCookie(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		session, err := app.Service.SessionService.GetSessionByID(cookie.Value)
		if err != nil {
			cookies.DeleteCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		if session.ExpireTime.Before(time.Now()) {
			cookies.DeleteCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		user, err := app.Service.UserService.GetUserByID(session.UID)
		if err != nil {
			cookies.DeleteCookie(w)
			app.Service.SessionService.DeleteSessionByID(cookie.Value)
			next.ServeHTTP(w, r)
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Middle) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		if (user == models.User{}) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Middle) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				logger.GetLogger().Error(err.(error).Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *Middle) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.GetLogger().Info(fmt.Sprintf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.RequestURI))
		next.ServeHTTP(w, r)
	})
}

func (app *Middle) SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}

func getUserFromContext(r *http.Request) models.User {
	user, ok := r.Context().Value(contextKeyUser).(models.User)
	if !ok {
		logger.GetLogger().Info("User is not authenticated")
		return models.User{}
	}
	return user
}
