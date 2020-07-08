package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	userstore "github.com/dalais/sdku_backend/store/user"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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

	cnf.StoreSession = sessions.NewCookieStore(cnf.Conf.APPKey)

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
	sa := sr.PathPrefix("/auth/").Subrouter()
	sa.Handle("/", components.UserJwtMiddleware.Handler(AuthValidate)).Methods("POST")
	sa.Handle("/register", auth.Registration()).Methods("POST")
	sa.Handle("/login", auth.Login()).Methods("POST")

	sr.Handle("/status", components.UserJwtMiddleware.Handler(StatusHandler)).Methods("GET")
	sr.Handle("/products", components.UserJwtMiddleware.Handler(
		producthandler.Index(),
	),
	).Methods("GET")

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

// AuthValidate ...
var AuthValidate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("access_token")
	token, _ := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	claims, _ := token.Claims.(jwt.MapClaims)
	data := claims["data"].(string)
	tokenData := components.TokenData{}
	jsData := components.DecryptStr(cnf.Conf.APPKey, data)
	json.Unmarshal([]byte(jsData), &tokenData)
	u := userstore.User{}
	row := store.Db.QueryRow(`SELECT user_id FROM auth_tokens WHERE id=$1`, tokenData.AuthID).Scan(&u.ID)
	if row != nil {
		fmt.Println(row)
	}
	row = store.Db.QueryRow(`SELECT id, name, role FROM users WHERE id=$1`, u.ID).Scan(
		&u.ID, &u.Name, &u.Role)
	if row != nil {
		fmt.Println(row)
	}
	answer := components.ReqAnswer{}
	answer.Data = u
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
})
