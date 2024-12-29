package middleware

import (
	"context"
	"database/sql"
	"errors"
	"goth/internal/service/jwthelper"
	"goth/internal/service/session"
	"goth/internal/store"
	"log/slog"
	"net/http"
)

type SessionMiddleware struct {
	GoogleUserStore store.GoogleUserStore
}

func NewSessionMiddleware(googleUserStore store.GoogleUserStore) *SessionMiddleware {
	return &SessionMiddleware{
		googleUserStore,
	}
}

// AddUserToContextMiddleware reads the 'auth_token' cookie, validates the JWT,
// and adds the userID to the context if the token is valid
func (s *SessionMiddleware) AddUserToContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var currentSession = session.Session{}

		// Read the auth_token cookie
		cookie, err := r.Cookie("auth_token")
		var errorNoCookiePresent = errors.Is(err, http.ErrNoCookie)
		if err != nil {
			if !errorNoCookiePresent {
				slog.Error(err.Error())
			}
			currentSession.IsDemo = true
			currentSession.CurrentUser = nil
		}

		if cookie == nil {
			next.ServeHTTP(w, r.WithContext(context.TODO()))
			return
		}

		// Validate the JWT token
		userID, err := jwthelper.ValidateJWT(cookie.Value)
		if err != nil {
			currentSession.IsDemo = true
			currentSession.CurrentUser = nil
		}

		if userID == "" {
			next.ServeHTTP(w, r.WithContext(context.TODO()))
		}

		user, err := s.GoogleUserStore.GetGoogleUserWhereGoogleId(userID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				panic(err)
			}
		} else {
			currentSession.IsDemo = false
			currentSession.CurrentUser = user
		}

		// Add the userID to the context
		ctx := context.WithValue(r.Context(), "session", &currentSession)

		// Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
