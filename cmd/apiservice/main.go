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

	"github.com/dalais/sdku_backend/chttp"
	gl "github.com/dalais/sdku_backend/cmd/global"
	"github.com/dalais/sdku_backend/config"
	"github.com/dalais/sdku_backend/handlers/auth"
	producthandler "github.com/dalais/sdku_backend/handlers/products"
	"github.com/dalais/sdku_backend/store"
	"github.com/gomodule/redigo/redis"
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

	sr := r.PathPrefix("/api/").Subrouter()

	/* csrfMiddleware := csrf.Protect(gl.Conf.CSRFKey, csrf.TrustedOrigins(gl.Conf.Front.Host))
	sr.Use(csrfMiddleware) */
	sa := sr.PathPrefix("/auth/").Subrouter()
	sa.Handle("/session", SessionValidate).Methods("POST")
	sa.Handle("/register", auth.Registration()).Methods("POST")
	sa.Handle("/login", auth.Login()).Methods("POST")

	sr.Handle("/status", authMdlw(
		authMdlw(StatusHandler),
	),
	).Methods("GET")

	sr.Handle("/products", authMdlw(
		producthandler.Index(),
	),
	).Methods("GET")

	// CORS
	http.ListenAndServe(":"+gl.Conf.Server.Port, handlers.CORS(
		handlers.AllowCredentials(),
		handlers.AllowedHeaders([]string{"X-Requested-With", "X-CSRF-Token", "Authorization", "Content-Type"}),
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

// SessionValidate ...
var SessionValidate = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authData := chttp.NewSessionData()
	//authData.CSRF = csrf.Token(r)

	session, _ := gl.StoreSession.Get(r, "sessid")

	if !session.IsNew && session.Values["auth"] == true {
		conn := gl.RPool.Get()
		defer conn.Close()
		// Get secret from redis
		redisToken := chttp.TokenRedis{}
		values, err := redis.Values(conn.Do("HGETALL", "token:"+fmt.Sprintf("%v", session.Values["token_id"])))
		if err == nil && len(values) > 0 {
			err := redis.ScanStruct(values, &redisToken)
			if err == nil {
				authData.Token = redisToken.Token
				authData.Role = redisToken.Role
				authData.UserID = redisToken.UserID
				authData.IsLogged = true
			} else {
				log.Printf("Error scanning the data token from the redis. Error: %v", err.Error())
			}
		}

		if authData.IsLogged == false {
			err := store.Db.QueryRow("SELECT access_token, user_id FROM auth_tokens WHERE id=$1", fmt.Sprintf("%v", session.Values["token_id"])).
				Scan(&authData.Token, &authData.UserID)
			if err != nil {
				log.Printf("Token not found, error: %v", err)
			}
			err = store.Db.QueryRow("SELECT role FROM users WHERE id=$1", fmt.Sprintf("%v", session.Values["user_id"])).Scan(&authData.Role)
			if err != nil {
				log.Printf("User not found, error: %v", err)
			}
			if err == nil {
				authData.IsLogged = true
			}
		}
	}

	//w.Header().Set("X-CSRF-Token", csrf.Token(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authData)
})

func authMdlw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt := chttp.JwtMdlw.Handler(next)
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "404 Not Found", http.StatusNotFound)
		} else {
			jwt.ServeHTTP(w, r)
		}
	})
}
