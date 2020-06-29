package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dalais/sdku_backend/config"
	"github.com/dalais/sdku_backend/handlers/auth"
	producthandler "github.com/dalais/sdku_backend/handlers/products"
	"github.com/dalais/sdku_backend/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

// TokenObj ...
type TokenObj struct {
	Token string `json:"token"`
}

// CustJwtMiddleware ...
type CustJwtMiddleware struct {
	*jwtmiddleware.JWTMiddleware
}

// APIKey ... Глобальный секретный ключ
var APIKey []byte
var conf config.LocalConfig

// init вызовется перед main()
func init() {
	file, err := os.Open("./config.yml")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	APIKey = []byte(conf.APPKey)

	// подключение к базе
	dbEngine := conf.Database.Connection
	dbURL := conf.Database.Connection + "://" + conf.Database.User + ":" + conf.Database.Pass + "@" + conf.Database.Host + "/" + conf.Database.Db

	var db *sql.DB
	db, err = sql.Open(dbEngine, dbURL)
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}

	// Настройка количества подключений к базе данных
	// для большей информации можно почитать https://www.alexedwards.net/blog/configuring-sqldb
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	models.Db = db
}

func main() {
	// Инициализируем gorilla/mux роутер
	r := mux.NewRouter()

	// Страница по умолчанию.
	r.Handle("/", http.FileServer(http.Dir("./views/")))

	sr := r.PathPrefix("/api/").Subrouter()
	sr.Handle("/auth/jwt", GetTokenHandler).Methods("GET")
	sr.Handle("/auth/register", auth.Register()).Methods("POST")
	sr.Handle("/status", StatusHandler).Methods("GET")
	sr.Handle("/products", jwtMiddleware.Handler(producthandler.Index())).Methods("GET")

	// Статику (картинки, скрипти, стили) будем раздавать
	// по определенному роуту /static/{file}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static/"))))
	// Наше приложение запускается на 8000 порту.
	// Для запуска мы указываем порт и наш роутер
	http.ListenAndServe(":"+conf.Server.Port, r)
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

// GetTokenHandler ...
var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter,
	r *http.Request) {
	// Создаем новый токен
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	// Устанавливаем набор параметров для токена
	claims["admin"] = true
	claims["name"] = "Adminushka"
	claims["exp"] = time.Now().Add(30 * 24 * time.Hour).Unix()

	// Подписываем токен нашим секретным ключем
	tokenString, _ := token.SignedString(APIKey)

	tokenObj := TokenObj{
		Token: tokenString,
	}
	// Отдаем токен клиенту
	addCookie(w, "access_token", tokenString, 1*time.Hour)
	w.Header().Set("Content-Type", "application/json")
	if string(APIKey) == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("API key is not found")
	} else {
		json.NewEncoder(w).Encode(tokenObj)
	}
})

var custJwtMiddle CustJwtMiddleware

// jwtMiddleware ...
var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	Extractor: custJwtMiddle.FromCookie,
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return APIKey, nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})

// FromCookie ... additional custom method for jwtmiddleware.JWTMiddleware
// which get token from cookie
func (jwtmiddleware CustJwtMiddleware) FromCookie(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Cookie")
	if authHeader == "" {
		return "", nil
	}
	cookie, _ := r.Cookie("access_token")
	if cookie == nil {
		return "", nil
	}

	return cookie.Value, nil
}

// addCookie will apply a new cookie to the response of a http request
// with the key/value specified.
func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		HttpOnly: true,
		Name:     name,
		Value:    value,
		Expires:  expire,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
}
