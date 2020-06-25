package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dalais/sdku_backend/config"
	"github.com/dalais/sdku_backend/handlers/auth"
	producthandler "github.com/dalais/sdku_backend/handlers/products"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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

// init вызовется перед main()
func init() {
	// Загрузка значений из .env файла в систему
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	conf := config.New()
	APIKey = []byte(conf.APPKey)
}

func main() {
	// Инициализируем gorilla/mux роутер
	r := mux.NewRouter()

	// Страница по умолчанию для нашего сайта это простой html.
	r.Handle("/", http.FileServer(http.Dir("./views/")))
	// Наше API состоит из трех роутов
	// /status - нужен для проверки работоспособности нашего API
	// /products - возвращаем набор продуктов,
	// по которым мы можем оставить отзыв
	// /products/{slug}/feedback - отображает фидбек пользователя по продукту
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
	http.ListenAndServe(":8000", r)
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

// FromCookie ...
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
