package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Mryashbhardwaj/marketAnalysis/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// user this file for handlers of auth bounded context

type authenticationMiddleware struct {
	logger              *slog.Logger
	db                  *pgxpool.Pool
	secretKey           string
	tokenInactivity     time.Duration
	tokenAbsoluteExpiry time.Duration
}

func GetAuthMiddleware(logger *slog.Logger, db *pgxpool.Pool, authConfig config.AuthConfig) *authenticationMiddleware {
	return &authenticationMiddleware{
		logger:              logger,
		db:                  db,
		secretKey:           authConfig.SecretKey,
		tokenInactivity:     time.Duration(authConfig.TokenInactivityMinutes) * time.Minute,
		tokenAbsoluteExpiry: time.Duration(authConfig.TokenAbsoluteExpiryMinutes) * time.Minute,
	}
}

func (amw *authenticationMiddleware) validateToken(token string) error {
	return nil
}

// Middleware function, which will be called for each request
func (amw *authenticationMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")

		if err := amw.validateToken(token); err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			// amw.logger.Info("Authenticated user %s\n", user)
			next.ServeHTTP(w, r)
		}
	})
}
