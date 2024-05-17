package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0" //app version

// config holds app configuration
type config struct {
	port int
	smtp struct { //SMTP credentials
		host     string
		port     int
		username string
		password string
	}
	frontend string
}

// application is the receiver for the various parts of the application
type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting invoice microservice on port %d", app.config.port)
	app.infoLog.Println(".....................................")
	return srv.ListenAndServe()
}

// main is the application entry point
func main() {
	var cfg config

	//Getting command line arguments
	flag.IntVar(&cfg.port, "port", 5000, "API Server port to listen on")
	flag.StringVar(&cfg.smtp.host, "smtphost", "live.smtp.mailtrap.io", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtpport", 2525, "smtp Server port to listen on")
	flag.StringVar(&cfg.smtp.username, "smtpusername", "api", "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtppassword", "6bf8fcf8a87e4b6cbb6e1cf1a8d0b4a3", "smtp password")
	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:4000", "frontend url")

	flag.Parse()

	//Getting environment variables
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
	}

	app.CreateDirIfNotExist("./invoices")
	err := app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
