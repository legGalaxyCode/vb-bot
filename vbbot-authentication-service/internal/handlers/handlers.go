package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"net/http"
	"vbbot-authentication-service/data"
	"vbbot-authentication-service/pkg/helpers"
	"vbbot-authentication-service/pkg/logging"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func (app *Config) Routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // default value
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	mux.Post("/authenticate", app.Authenticate)

	return mux
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	logger.Println("start authentication")
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := helpers.ReadJSON(w, r, &requestPayload)
	if err != nil {
		logger.Println("error read json")
		helpers.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	logger.Println(requestPayload)

	// validate the user against database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		logger.Printf("error get app.Model.User by email: %s\n", err)
		helpers.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		logger.Println("error in validation")
		helpers.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}
	helpers.WriteJSON(w, http.StatusAccepted, payload)
}
