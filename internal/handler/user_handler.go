package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"golang-web/internal/auth"
	"golang-web/internal/model"
	"golang-web/internal/service"
	"golang-web/web"

	"github.com/go-playground/form/v4"
	"github.com/sirupsen/logrus"
)

// availableRoles is defined here (removed from var block below if duplicate)

type UserHandler struct {
	Log         *logrus.Logger
	Renderer    *web.Renderer
	Decoder     *form.Decoder
	UserService *service.UserService
}

func NewUserHandler(logger *logrus.Logger, renderer *web.Renderer, decoder *form.Decoder, userService *service.UserService) *UserHandler {
	return &UserHandler{
		Log:         logger,
		Renderer:    renderer,
		Decoder:     decoder,
		UserService: userService,
	}
}

var availableRoles = []string{auth.RoleOwner, auth.RoleStaff}

// CreateUserForm renders the create-user form page (GET /users/create).
func (h *UserHandler) CreateUserForm(w http.ResponseWriter, r *http.Request) {
	userForm := model.FormData[model.CreateUserRequest]{
		Judul:  "Tambah Pengguna",
		Action: "/users",
		Data:   model.CreateUserRequest{},
		Errors: nil,
	}

	data := map[string]any{
		"title":    "Tambah Pengguna",
		"active":   "users",
		"roles":    availableRoles,
		"isEdit":   false,
		"UserForm": userForm,
	}

	h.Renderer.Render(w, "users/form.html", data)
}

// EditUserForm renders the edit-user form page (GET /users/{id}/edit).
func (h *UserHandler) EditUserForm(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.GetUserByID(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}

	userForm := model.FormData[model.UpdateUserRequest]{
		Judul:  "Ubah Pengguna",
		Action: fmt.Sprintf("/users/%d/edit", ID),
		Data: model.UpdateUserRequest{
			Name: user.Name,
			Role: user.Role,
		},
		Errors: nil,
	}

	data := map[string]any{
		"title":    "Ubah Pengguna",
		"active":   "users",
		"roles":    availableRoles,
		"isEdit":   true,
		"UserForm": userForm,
	}

	h.Renderer.Render(w, "users/form.html", data)
}

// ChangePasswordForm renders the change-password form page (GET /users/{id}/password).
func (h *UserHandler) ChangePasswordForm(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.GetUserByID(r.Context(), ID)
	if err != nil {
		http.Error(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}

	userForm := model.FormData[model.ChangePasswordRequest]{
		Judul:  "Ganti Password",
		Action: fmt.Sprintf("/users/%d/password", ID),
		Data:   model.ChangePasswordRequest{},
		Errors: nil,
	}

	data := map[string]any{
		"title":    "Ganti Password",
		"active":   "users",
		"User":     user,
		"UserForm": userForm,
	}

	h.Renderer.Render(w, "users/password.html", data)
}

// GetAllUsers renders the user list page.
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.GetAllUsers(r.Context())
	if err != nil {
		h.Log.WithError(err).Error("error get users")
		http.Error(w, "Gagal mengambil data pengguna", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"title":  "Manajemen Pengguna",
		"active": "users",
		"users":  users,
	}

	h.Renderer.Render(w, "users/index.html", data)
}

// CreateUser handles the POST /users form submission.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.CreateUserRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	_, formError, err := h.UserService.CreateUser(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error create user")
		http.Error(w, "Gagal membuat pengguna", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		userForm := model.FormData[model.CreateUserRequest]{
			Judul:  "Tambah Pengguna",
			Action: "/users",
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":    "Tambah Pengguna",
			"active":   "users",
			"roles":    availableRoles,
			"isEdit":   false,
			"UserForm": userForm,
		}

		h.Renderer.Render(w, "users/form.html", data)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// EditUser handles the POST /users/{id}/edit form submission.
func (h *UserHandler) EditUser(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.UpdateUserRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	request.ID = ID

	_, formError, err := h.UserService.UpdateUser(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error update user")
		http.Error(w, "Gagal mengubah pengguna", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		userForm := model.FormData[model.UpdateUserRequest]{
			Judul:  "Ubah Pengguna",
			Action: fmt.Sprintf("/users/%d/edit", ID),
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":    "Ubah Pengguna",
			"active":   "users",
			"roles":    availableRoles,
			"isEdit":   true,
			"UserForm": userForm,
		}

		h.Renderer.Render(w, "users/form.html", data)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// ChangePassword handles POST /users/{id}/password.
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request := new(model.ChangePasswordRequest)
	if err := h.Decoder.Decode(request, r.PostForm); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	request.ID = ID

	formError, err := h.UserService.ChangePassword(r.Context(), request)
	if err != nil {
		h.Log.WithError(err).Error("error change password")
		http.Error(w, "Gagal mengubah password", http.StatusInternalServerError)
		return
	}

	if formError != nil {
		user, _ := h.UserService.GetUserByID(r.Context(), ID)

		userForm := model.FormData[model.ChangePasswordRequest]{
			Judul:  "Ganti Password",
			Action: fmt.Sprintf("/users/%d/password", ID),
			Data:   *request,
			Errors: formError,
		}

		data := map[string]any{
			"title":    "Ganti Password",
			"active":   "users",
			"User":     user,
			"UserForm": userForm,
		}

		h.Renderer.Render(w, "users/password.html", data)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// DeleteUser handles POST /users/{id}/delete.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("id")

	ID, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	request := &model.DeleteUserRequest{ID: ID}

	if err := h.UserService.DeleteUser(r.Context(), request); err != nil {
		h.Log.WithError(err).Error("error delete user")
		http.Error(w, "Gagal menghapus pengguna", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}
