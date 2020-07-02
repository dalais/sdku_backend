package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dalais/sdku_backend/cmd/cnf"
	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/config"
	"github.com/dalais/sdku_backend/handlers/auth"
	producthandler "github.com/dalais/sdku_backend/handlers/products"
	"github.com/dalais/sdku_backend/store"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// init вызовется перед main()
func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	cnf.ROOT = dir

	// Load environment variables from .env
	if err := godotenv.Load(cnf.ROOT + "/.env"); err != nil {
		log.Print("No .env file found")
	}

	cnf.Conf = *config.New()

	cnf.APIKey = []byte(cnf.Conf.APPKey)

	// Strings for database connection
	dbEngine := cnf.Conf.Database.Connection
	dbURL := cnf.Conf.Database.Connection + "://" +
		cnf.Conf.Database.User + ":" +
		cnf.Conf.Database.Pass + "@" +
		cnf.Conf.Database.Host + "/" +
		cnf.Conf.Database.Db

	// Database connection
	cnf.Db, err = sql.Open(dbEngine, dbURL)
	if err != nil {
		log.Panic(err)
	}

	if err = cnf.Db.Ping(); err != nil {
		log.Panic(err)
	}

	// Configuring the number of database connections
	// for more information https://www.alexedwards.net/blog/configuring-sqldb
	cnf.Db.SetMaxOpenConns(25)
	cnf.Db.SetMaxIdleConns(25)
	cnf.Db.SetConnMaxLifetime(5 * time.Minute)
	store.Db = cnf.Db
}

func main() {
	// Init router
	r := mux.NewRouter()

	// Default page.
	r.Handle("/", http.FileServer(http.Dir(cnf.ROOT+"/views/")))

	sr := r.PathPrefix("/api/").Subrouter()
	sr.Handle("/auth/jwt", components.GetTokenHandler).Methods("GET")
	sr.Handle("/auth/register", auth.Registration()).Methods("POST")
	sr.Handle("/auth/login", auth.Login()).Methods("POST")
	sr.Handle("/status", StatusHandler).Methods("GET")
	sr.Handle("/products", components.MyJwtMiddleware.Handler(producthandler.Index())).Methods("GET")

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir(cnf.ROOT+"/static/"))))

	http.ListenAndServe(":"+cnf.Conf.Server.Port, r)
}

// NotImplemented ...
var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not Implemented"))
})

// StatusHandler ... The status handler will be invoked when the user calls the /status route
//  It will simply return a string with the message "API is up and running"
var StatusHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
})
