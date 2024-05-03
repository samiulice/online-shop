package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"online_store/internal/driver"
	"online_store/internal/repository"
	"online_store/internal/repository/dbrepo"
	"os"
	"time"
)

const version = "1.0.0" //app version

// config holds app configuration
type config struct {
	port int
	env  string //production or development mode
	db   struct {
		dsn string //Data source name : database connection name
	}
	stripe struct {
		secret string //secret key for privacy purpose
		key    string //Publishable key
	}

	smtp struct { //SMTP credentials
		host     string
		port     int
		username string
		password string
	}
	secretKey string
	frontend  string
}

// application is the receiver for the various parts of the application
type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       repository.DatabaseRepo
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

	app.infoLog.Printf("Starting HTTP Back end server in %s mode on port %d", app.config.env, app.config.port)
	app.infoLog.Println(".....................................")
	return srv.ListenAndServe()
}

// main is the application entry point
func main() {
	var cfg config

	//Getting command line arguments
	flag.IntVar(&cfg.port, "port", 4001, "API Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application Environment{development|production}")
	flag.StringVar(&cfg.db.dsn, "dsn", "host=localhost port=5432 dbname=online_store user=postgres password=samiul@10526 sslmode=disable", "DSN")
	flag.StringVar(&cfg.smtp.host, "smtphost", "live.smtp.mailtrap.io", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtpport", 2525, "smtp Server port to listen on")
	flag.StringVar(&cfg.smtp.username, "smtpusername", "api", "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtppassword", "6bf8fcf8a87e4b6cbb6e1cf1a8d0b4a3", "smtp password")
	flag.StringVar(&cfg.secretKey, "secretkey", "Oanlsm1SeiEti25SL1iuVSunr06LOmeo", "secret key")
	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:4000", "frontend url")

	flag.Parse()

	//Getting environment variables
	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//Connection to database
	dbConn, err := driver.ConnectDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
		return
	}
	defer dbConn.Close()
	db := dbrepo.NewDBRepo(dbConn)
	infoLog.Println("Connected to database")

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       db,
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
