package main

import (
	"fmt"
	"log"
	"net/http"
	"vbbot-api-service/internal/server"
	"vbbot-api-service/pkg/logging"
)

const webPort = "8001"

func main() {
	logging.Init()
	logger := logging.GetLogger()
	app := server.Server{
		Logger: logger,
	}

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	fmt.Println("Start server on port ", webPort)
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
