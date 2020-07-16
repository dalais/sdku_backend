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

	gl "github.com/dalais/sdku_backend/cmd/global"
	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/config"
	"github.com/dalais/sdku_backend/handlers/auth"
	producthandler "github.com/dalais/sdku_backend/handlers/products"
	"github.com/dalais/sdku_backend/store"
	userstore "github.com/dalais/sdku_backend/store/user"
	"github.com/gorilla/handlers"
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
	gl.ROOT = dir

	// Load environment variables from .env
	if err := godotenv.Load(gl.ROOT + "/.env"); err != nil {
		log.Print("No .env file found")
	}

	gl.Conf = *config.New()

	gl.StoreSession = sessions.NewCookieStore(gl.Conf.APPKey)

	// Strings for database connection
	dbEngine := gl.Conf.Database.Connection
	dbURL := gl.Conf.Database.Connection + "://" +
		gl.Conf.Database.User + ":" +
		gl.Conf.Database.Pass + "@" +
		gl.Conf.Database.Host + "/" +
		gl.Conf.Database.Db

	// Database connection
	gl.Db, err = sql.Open(dbEngine, dbURL)
	if err != nil {
		log.Panic(err)
	}

	if err = gl.Db.Ping(); err != nil {
		log.Panic(err)
	}

	// Configuring the number of database connections
	// for more information https://www.alexedwards.net/blog/configuring-sqldb
	gl.Db.SetMaxOpenConns(25)
	gl.Db.SetMaxIdleConns(25)
	gl.Db.SetConnMaxLifetime(5 * time.Minute)
	store.Db = gl.Db

	// Redis connection
	gl.InitRPool()
	err = gl.Rping()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	// Init router
	r := mux.NewRouter()

	// Default page.
	r.Handle("/", http.FileServer(http.Dir(gl.ROOT+"/views/")))

	sr := r.PathPrefix("/api/").Subrouter()

	sa := sr.PathPrefix("/auth/").Subrouter()
	sa.Handle("/verify", components.JwtMdlw.Handler(AuthValidate)).Methods("GET")
	sa.Handle("/register", auth.Registration()).Methods("POST")
	sa.Handle("/login", auth.Login()).Methods("POST")

	sr.Handle("/status", sessionMdlw(
		components.JwtMdlw.Handler(StatusHandler),
	),
	).Methods("GET")

	sr.Handle("/products", components.JwtMdlw.Handler(
		producthandler.Index(),
	),
	).Methods("GET")

	// CORS
	http.ListenAndServe(":"+gl.Conf.Server.Port, handlers.CORS(
		handlers.AllowCredentials(),
		handlers.AllowedHeaders([]string{"X-Requested-With", " X-HTTP-Method-Override", "Content-Type"}),
		handlers.AllowedOrigins(gl.Conf.Front.Host),
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "OPTIONS"}),
	)(r))
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
	cookie, _ := r.Cookie("_token")
	tokenData := components.TokenPayload(cookie.Value)
	u := userstore.User{}
	row := store.Db.QueryRow(`SELECT id, name, role FROM users WHERE id=$1`, tokenData.UserID).Scan(
		&u.ID, &u.Name, &u.Role)
	if row != nil {
		fmt.Println(row)
	}
	answer := components.ReqAnswer{}
	answer.Data = u
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
})

func sessionMdlw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := gl.StoreSession.Get(r, "sessid")
		authHeader := r.Header.Get("Cookie")
		cookie, _ := r.Cookie("_token")
		if session.IsNew {
			c := &http.Cookie{
				Name:    "_token",
				Value:   "",
				Path:    "/",
				Expires: time.Unix(0, 0),

				HttpOnly: true,
			}
			http.SetCookie(w, c)
			if authHeader != "" && cookie != nil {
				tokenData := components.TokenPayload(cookie.Value)
				_ = gl.Db.QueryRow(`DELETE FROM auth_access WHERE token_id=$1`, tokenData.TokenID)

			}
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
