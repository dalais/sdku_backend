package chttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	gl "github.com/dalais/sdku_backend/cmd/global"
	"github.com/dalais/sdku_backend/store"
	"github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
)

// TokenObj ...
type TokenObj struct {
	ID      int64  `json:"id"`
	Token   string `json:"token"`
	LifeSec int    `json:"life_sec"`
	Secret  string `json:"secret"`
}

// NewTokenObj ...
func NewTokenObj() TokenObj {
	tObj := TokenObj{}
	tObj.ID = 0
	tObj.LifeSec = 43200
	tObj.Token = ""
	tObj.Secret = ""
	return tObj
}

// TokenData ...
type TokenData struct {
	TokenID int64 `json:"token_id"`
	UserID  int64 `json:"user_id"`
}

// CustJwtMiddleware ...
type CustJwtMiddleware struct {
	*jwtmiddleware.JWTMiddleware
}

// GetToken for login proccess
var GetToken = func(key string, tokenID int64, userID int64) TokenObj {
	// Create new token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	tokenData := TokenData{}
	tokenData.TokenID = tokenID
	tokenData.UserID = userID
	jsData, _ := json.Marshal(tokenData)
	strData := string(jsData)
	encrytedData := EncryptStr(gl.Conf.APPKey, strData)
	// Set params for payload
	claims["data"] = encrytedData
	claims["exp"] = time.Now().Add(30 * 24 * time.Hour).Unix()

	// Signing the token
	tokenString, _ := token.SignedString([]byte(key))

	tokenObj := NewTokenObj()
	tokenObj.ID = tokenID
	tokenObj.Token = tokenString
	tokenObj.Secret = key
	return tokenObj
}

var custJwtMiddle CustJwtMiddleware

// JwtMdlw ...
var JwtMdlw = jwtmiddleware.New(jwtmiddleware.Options{
	Extractor: custJwtMiddle.FromCookie,
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		var secret []byte
		claims := token.Claims.(jwt.MapClaims)
		data := claims["data"].(string)
		tokenData := TokenData{}
		jsData := DecryptStr(gl.Conf.APPKey, data)
		json.Unmarshal([]byte(jsData), &tokenData)

		conn := gl.RPool.Get()
		defer conn.Close()
		// Get secret from redis
		redisSecret, err := redis.String(conn.Do("HGET", "access:"+fmt.Sprintf("%v", tokenData.TokenID), "secret"))
		if err == nil || err != redis.ErrNil {
			secret = []byte(redisSecret)
		} else {
			// Select secret from db
			row := store.Db.QueryRow(`SELECT secret FROM auth_access WHERE token_id=$1`, tokenData.TokenID).Scan(&secret)
			if row != nil {
				fmt.Println(row)
			}
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
	cookie, _ := r.Cookie("_token")
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

// TokenPayload - get token from cookie and parse data
func TokenPayload(cookieVal string) TokenData {
	token, _ := jwt.Parse(cookieVal, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	claims, _ := token.Claims.(jwt.MapClaims)
	data := claims["data"].(string)
	tokenData := TokenData{}
	jsData := DecryptStr(gl.Conf.APPKey, data)
	json.Unmarshal([]byte(jsData), &tokenData)

	return tokenData
}
