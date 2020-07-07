package components

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dalais/sdku_backend/cmd/cnf"
	"github.com/dalais/sdku_backend/store"
	"github.com/dgrijalva/jwt-go"
)

// TokenObj ...
type TokenObj struct {
	Token string `json:"token"`
}

// TokenData ...
type TokenData struct {
	AuthID int64 `json:"auth_id"`
}

// CustJwtMiddleware ...
type CustJwtMiddleware struct {
	*jwtmiddleware.JWTMiddleware
}

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
	tokenString, _ := token.SignedString(cnf.Conf.APPKey)

	tokenObj := TokenObj{
		Token: tokenString,
	}
	// Отдаем токен клиенту
	SendTokenToCookie(w, "access_token", tokenString, 1*time.Hour)
	if string(cnf.Conf.APPKey) == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("API key is not found")
	} else {
		json.NewEncoder(w).Encode(tokenObj)
	}
})

// GetToken for login proccess
var GetToken = func(key string, tokenID int64) TokenObj {
	// Create new token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	tokenData := TokenData{}
	tokenData.AuthID = tokenID
	jsData, _ := json.Marshal(tokenData)
	strData := string(jsData)
	encrytedData := EncryptStr(cnf.Conf.APPKey, strData)
	// Set params for payload
	claims["data"] = encrytedData
	claims["exp"] = time.Now().Add(30 * 24 * time.Hour).Unix()

	// Signing the token
	tokenString, _ := token.SignedString([]byte(key))

	tokenObj := TokenObj{
		Token: tokenString,
	}
	return tokenObj
}

var custJwtMiddle CustJwtMiddleware

// AppJwtMiddleware ...
var AppJwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	Extractor: custJwtMiddle.FromCookie,
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return cnf.Conf.APPKey, nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})

// UserJwtMiddleware ...
var UserJwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	Extractor: custJwtMiddle.FromCookie,
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		var secret []byte
		claims := token.Claims.(jwt.MapClaims)
		data := claims["data"].(string)
		tokenData := TokenData{}
		jsData := DecryptStr(cnf.Conf.APPKey, data)
		json.Unmarshal([]byte(jsData), &tokenData)
		// Select secret from db
		row := store.Db.QueryRow(`SELECT secret FROM auth_access WHERE token_id=$1`, tokenData.AuthID).Scan(&secret)
		if row != nil {
			fmt.Println(row)
		}
		return secret, nil
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

// SendTokenToCookie will apply a new cookie to the response of a http request
// with the key/value specified.
func SendTokenToCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
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
