package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"online_store/internal/driver"
	"online_store/internal/repository"
	"online_store/internal/repository/dbrepo"
	"os"
	"time"
)

const version = "1.0.0" //app version
const cssVersion = "1"  //cssVersion informed the browser about the correct css version

// config holds app configuration
type config struct {
	port int
	env  string //production or development mode
	api  string //URL that will be called for backend api
	db   struct {
		dsn string //Data source name : database connection name
	}
	stripe struct {
		secret string //secret key for privacy purpose
		key    string //Publishable key
	}
}

// application is the receiver for the various parts of the application
type application struct {
	config       config
	infoLog      *log.Logger
	errorLog     *log.Logger
	temlateCache map[string]*template.Template
	version      string
	DB           repository.DatabaseRepo
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

	app.infoLog.Printf("Starting HTTP server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()
}

// main is the application entry point
func main() {
	var cfg config

	//Getting command line arguments
	flag.IntVar(&cfg.port, "port", 4000, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application Environment{development|production}")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "URL to api")
	flag.StringVar(&cfg.db.dsn, "dsn", "host=localhost port=5432 dbname=online_store user=postgres password=samiul@10526 sslmode=disable", "DSN")

	flag.Parse()

	//Getting environment variables
	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//Connection to database
	dbConn, err := driver.ConnectDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatalln(err)
		return
	}
	defer dbConn.Close()
	infoLog.Println("Connected to database")

	db := dbrepo.NewDBRepo(dbConn)

	tc := make(map[string]*template.Template)

	app := &application{
		config:       cfg,
		infoLog:      infoLog,
		errorLog:     errorLog,
		temlateCache: tc,
		version:      version,
		DB:           db,
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}

}
