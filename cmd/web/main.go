package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"online_store/internal/driver"
	"online_store/internal/models"
	"online_store/internal/repository"
	"online_store/internal/repository/dbrepo"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const version = "1.0.0" //app version
const cssVersion = "1"  //cssVersion informed the browser about the correct css version
var session *scs.SessionManager

// config holds app configuration
type config struct {
	host string //localhost or remoteserver
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
	secretKey string
	frontend  string
}

// application is the receiver for the various parts of the application
type application struct {
	config       config
	infoLog      *log.Logger
	errorLog     *log.Logger
	temlateCache map[string]*template.Template
	version      string
	DB           repository.DatabaseRepo
	Session      *scs.SessionManager
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", app.config.host, app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting HTTP server in %s mode on port %d", app.config.env, app.config.port)
	app.infoLog.Println(".....................................")

	return srv.ListenAndServe()
}

// main is the application entry point
func main() {
	//Register the types of sessional variable
	gob.Register(models.TransactionData{})
	gob.Register(models.User{})

	//app config
	var cfg config

	//Getting command line arguments
	flag.StringVar(&cfg.host, "host", "localhost", "Server port to listen on")
	flag.IntVar(&cfg.port, "port", 4000, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application Environment{development|production}")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "URL to api")
	flag.StringVar(&cfg.db.dsn, "dsn", "host=localhost port=5432 dbname=online_store user=postgres password=samiul@10526 sslmode=disable", "DSN")
	flag.StringVar(&cfg.secretKey, "secretkey", "Oanlsm1SeiEti25SL1iuVSunr06LOmeo", "secret key of length 32chars") //
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
		errorLog.Fatalln(err)
		return
	}
	defer dbConn.Close()

	db := dbrepo.NewDBRepo(dbConn)
	infoLog.Println("Connected to database")

	// Establish connection pool to PostgreSQL.
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:samiul@10526@localhost/online_store")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	//set up session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true //change it to false if it needs to delete cookie at the closing of the browser
	session.Cookie.Secure = false //localhost is insecure connection which is used in InProduction mode
	session.Store = pgxstore.New(pool)

	tc := make(map[string]*template.Template)

	app := &application{
		config:       cfg,
		infoLog:      infoLog,
		errorLog:     errorLog,
		temlateCache: tc,
		version:      version,
		DB:           db,
		Session:      session,
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
