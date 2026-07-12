package main

import (
	_ "Gateway/docs"
	"log/slog"
	"net/http"
	"os"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @title Swagger Example API
// @version 2.0
// @description Gateway
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host
func main() {
	slog.Info("Gateway started")

	appSettings := settings{
		microserviceURL: os.Getenv("MICROSERVICE_URL"),
		userServiceURL:  os.Getenv("USER_SERVICE_URL"),
	}

	server := &http.Server{
		Addr:    ":8083",
		Handler: appSettings.routes(),
	}

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("gateway server start error:", err)
	}
}
