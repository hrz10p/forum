package main

import (
	"database/sql"
	"fmt"
	"forum/pkg/services"
	"forum/pkg/utils/logger"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Application struct {
	Service *services.Service
	Router  *http.ServeMux
	Logger  *logger.Logger
	Config  *Config
}

// Config struct to hold application configuration
type Config struct { // Add configuration fields here
}

// NewApplication initializes a new Application struct
func NewApplication(config *Config) *Application {
	// Initialize services
	db, err := sql.Open("sqlite3", "forum.sqlite")
	if err != nil {
		fmt.Print(err)
	}

	// Initialize router
	router := http.NewServeMux()

	return &Application{
		Service: services.NewService(db),
		Router:  router,
		Logger:  logger.GetLogger(),
		Config:  config,
	}
}

// Start starts the application server
func (app *Application) Start(addr string) error {
	// Start the server
	app.Logger.Info("serv")
	app.InitializeRoutes()
	return http.ListenAndServe(addr, app.Router)
}
