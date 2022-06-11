package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"net/http"
	"vbbot-api-service/internal/handlers"
	"vbbot-api-service/pkg/logging"
)

type Server struct {
	Logger logging.Logger
}

func (srv *Server) Routes() http.Handler {
	mux := chi.NewRouter()

	// who is allowed to connect
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // default value
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	mux.Post("/handle", handlers.HandleSubmission)

	return mux
}
