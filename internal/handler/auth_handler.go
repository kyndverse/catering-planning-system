package handler

import (
	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/web"
	"net/http"

	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	Log         *logrus.Logger
	Renderer    *web.Renderer
	Decoder     *form.Decoder
	AuthService *service.AuthService
}

func NewAuthHandler(logger *logrus.Logger, renderer *web.Renderer, decoder *form.Decoder, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		Log:         logger,
		Renderer:    renderer,
		Decoder:     decoder,
		AuthService: authService,
	}
}

func (h *AuthHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	loginForm := model.FormData[model.LoginRequest]{
		Judul:  "Login",
		Action: "/auth/login",
		Data:   model.LoginRequest{},
		Errors: nil,
	}

	data := map[string]any{
		"title":      "Login",
		"IsAuthPage": true,
		"LoginForm":  loginForm,
	}

	h.Renderer.Render(w, "auth/login.html", data)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.LoginRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	var oldSessionID string
	cookie, err := r.Cookie("session_id")
	if err == nil {
		oldSessionID = cookie.Value
	}

	sessionID, formError, err := h.AuthService.Login(r.Context(), request, oldSessionID)
	if err != nil {
		h.Log.WithError(err).Error("error to login")
		http.Error(w, "Login Failed", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		loginForm := model.FormData[model.LoginRequest]{
			Judul:  "Login",
			Action: "/auth/login",
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":      "Login",
			"IsAuthPage": true,
			"LoginForm": loginForm,
		}

		h.Renderer.Render(w, "auth/login.html", data)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/ingredients", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var sessionID string
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionID = cookie.Value
	}

	if err := h.AuthService.Logout(r.Context(), sessionID); err != nil {
		h.Log.WithError(err).Error("error to login")
		http.Error(w, "Logout Failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}
