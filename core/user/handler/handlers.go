package handlers

import (
	"context"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Mryashbhardwaj/marketAnalysis/internal/utils"
)

type UserService interface {
	Login(ctx context.Context, email, password string) (string, error)
	Signup(name, email, password, broker string, photoFile multipart.File) error
	Logout(email string) error
}

type Handler struct {
	logger      *slog.Logger
	userService UserService
}

func GetHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	h.logger.Info("Login attempt", slog.String("email", email))

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	jwtToken, err := h.userService.Login(ctx, email, password)
	if err != nil {
		h.logger.Error("Login failed", slog.String("error", err.Error()))
		http.Error(w, "Unauthorized: Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
	})
}

func (h Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	broker := r.FormValue("broker")

	if name == "" || email == "" || password == "" || broker == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	file, _, err := r.FormFile("photo")

	err = h.userService.Signup(name, email, password, broker, file)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error signing up user: "+err.Error())
		return
	}

	jwtToken, err := h.userService.Login(email, password)
	if err != nil {
		h.logger.Error("Signup failed", slog.String("error", err.Error()))
		http.Error(w, "Unauthorized: Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
	})

	// var photo

	// newUser := User{Name: name, Email: email, Broker: broker, Photo: photoPath}
	utils.RespondWithJSON(w, http.StatusCreated, nil)

}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")

	err := h.userService.Logout(email)
	if err != nil {
		h.logger.Error("Logout failed", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{Name: "auth_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true}

	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
