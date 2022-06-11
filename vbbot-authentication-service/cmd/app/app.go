package main

import (
	"fmt"
	"net/http"
	"vbbot-authentication-service/data"
	"vbbot-authentication-service/internal/config"
	"vbbot-authentication-service/internal/handlers"
	"vbbot-authentication-service/pkg/logging"
)

const webPort = "80"

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Println("Setup connection")

	conn := config.ConnectToDB()
	if conn == nil {
		logger.Panic("error setup connection")
	}
	app := handlers.Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		logger.Panic(err)
	}
}
