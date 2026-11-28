package middleware

import (
	"errors"
	"golang-web/internal/auth"
	"golang-web/internal/repository"
	"golang-web/web"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	Log               *logrus.Logger
	SessionRepository *repository.SessionRepository
	Renderer          *web.Renderer
}

func NewAuthMiddleware(logger *logrus.Logger, sessionRepository *repository.SessionRepository, renderer *web.Renderer) *AuthMiddleware {
	return &AuthMiddleware{
		Log:               logger,
		SessionRepository: sessionRepository,
		Renderer:          renderer,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		session, err := m.SessionRepository.FindByID(r.Context(), sessionCookie.Value)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			m.Log.WithError(err).Error("failed to find session")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ctx := auth.WithUserID(r.Context(), session.UserID)
		ctx = auth.WithRole(ctx, session.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) Authorize(allowedRoles ...string) func(http.Handler) http.Handler {
	roles := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		roles[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := auth.RoleFromContext(r.Context())
			if !ok {
				m.forbidden(w)
				return
			}

			if _, allowed := roles[role]; !allowed {
				m.forbidden(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) forbidden(w http.ResponseWriter) {
	data := map[string]any{
		"title":  "Akses Ditolak",
		"active": "",
	}

	w.WriteHeader(http.StatusForbidden)
	m.Renderer.Render(w, "error/forbidden.html", data)
}
