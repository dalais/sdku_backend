package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dalais/sdku_backend/chttp"
	gl "github.com/dalais/sdku_backend/cmd/global"
	"github.com/dalais/sdku_backend/store"
	userstore "github.com/dalais/sdku_backend/store/user"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	custErrMsg = "Incorrect login or password"
)

// LoginError ...
type LoginError struct{}

func (m *LoginError) Error() string {
	return custErrMsg
}

// Login ...
func Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u *userstore.User
		var answer *chttp.ReqAnswer
		// Handle http request
		answer = chttp.PostReqHandler(&u, w, r)

		user := userstore.User{
			Email:      u.Email,
			Password:   u.Password,
			Role:       u.Role,
			RememberMe: u.RememberMe,
		}

		// Validation fields
		user, answer = loginValidation(user, answer)

		var token chttp.TokenObj

		// Create and store token
		if answer.IsEmptyError() {
			token = storeToken(user, answer)
		}

		// If everything is in order we set a notification and send the token to cookies
		if answer.IsEmptyError() {
			session, _ := gl.StoreSession.New(r, "sessid")
			session.Values["auth"] = true
			session.Values["user_id"] = user.ID
			session.Values["token_id"] = token.ID
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   0,
				HttpOnly: true,
			}
			if *user.RememberMe == true {
				session.Values["remember_me"] = user.RememberMe
				session.Options = &sessions.Options{
					Path:     "/",
					MaxAge:   86400 * 7,
					HttpOnly: true,
				}
			}
			err := session.Save(r, w)
			if err != nil {
				chttp.HandleAnswerError(err, answer, "Internal Server Error")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if answer.IsEmptyError() {
				answer.Message = "Authentication is successful"
				sessionData := chttp.NewSessionData()
				sessionData.IsLogged = true
				sessionData.UserID = user.ID
				sessionData.Token = token.Token
				sessionData.Role = user.Role
				answer.Data = append(answer.Data, sessionData)
			}
		}

		json.NewEncoder(w).Encode(answer)
	})
}

// LoginValidation ...
func loginValidation(user userstore.User, answer *chttp.ReqAnswer) (
	userstore.User, *chttp.ReqAnswer) {
	if answer.Error == 0 {
		// new error struct for answer.ErrMesgs
		errMsgs := struct {
			Error string `json:"error"`
		}{}
		// New answer struct
		answer = &chttp.ReqAnswer{}

		// Getting the password sent in the request
		password := []byte(user.Password)

		// Select user from db
		row := store.Db.QueryRow(`SELECT id, email, role, password, crtd_at FROM users WHERE email=$1`, user.Email).Scan(
			&user.ID, &user.Email, &user.Role, &user.Password, &user.CrtdAt)
		if row != nil {
			chttp.HandleAnswerError(row, answer, custErrMsg)
			answer.Data = nil
		}

		// If record exist, compare passwords
		if row == nil {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), password)
			chttp.HandleAnswerError(err, answer, custErrMsg)
		}
		user.Password = ""

		// Errors handling
		if errMsgs.Error != "" {
			chttp.HandleAnswerError(&LoginError{}, answer, custErrMsg)
		}

		// If there are no errors, we set user data for the answer.Data field
		if len(answer.ErrMesgs) == 0 {
			answer.Data = append(answer.Data, user)
		}
	}

	return user, answer
}

// TODO ...
func storeToken(user userstore.User, answer *chttp.ReqAnswer) chttp.TokenObj {

	var tokenID int64
	// Create sercret string
	secret := chttp.RandomString(32)

	// BEGIN transaction
	tx, err := store.Db.Begin()
	chttp.HandleAnswerError(err, answer, custErrMsg)

	createTokensql := `
				INSERT INTO auth_tokens (user_id)
					VALUES ($1) RETURNING id`
	err = tx.QueryRow(createTokensql, user.ID).Scan(&tokenID)
	if err != nil {
		tx.Rollback()
		chttp.HandleAnswerError(err, answer, custErrMsg)
	}

	// insert record into table2, referencing the first record from table1
	_, err = tx.Exec("INSERT INTO auth_access(token_id, secret) VALUES($1, $2)", tokenID, secret)
	if err != nil {
		tx.Rollback()
		chttp.HandleAnswerError(err, answer, custErrMsg)
	}
	// Create token object
	token := chttp.GetToken(secret, tokenID, user.ID)

	// insert access_token
	_, err = tx.Exec("UPDATE auth_tokens SET access_token=$1 WHERE id=$2", token.Token, tokenID)
	if err != nil {
		tx.Rollback()
		chttp.HandleAnswerError(err, answer, custErrMsg)
	}

	// COMMIT the transaction
	chttp.HandleAnswerError(tx.Commit(), answer, custErrMsg)
	if !answer.IsEmptyError() {
		token.Token = ""
		answer.Data = nil
	}

	tokenHMSet(user, token)

	return token
}

// Store token in redis
func tokenHMSet(user userstore.User, token chttp.TokenObj) {
	// get conn and put back when exit from method
	conn := gl.RPool.Get()
	defer conn.Close()

	tokenIDStr := fmt.Sprintf("%v", token.ID)

	_, err := conn.Do("HMSET", "token:"+tokenIDStr, "user_id", user.ID, "role", user.Role, "token", token.Token)
	if err != nil {
		log.Printf("ERROR: fail set key %s, val %s, error %s", "token:"+tokenIDStr, token.Token, err.Error())
	} else {
		_, err = conn.Do("EXPIRE", "token:"+tokenIDStr, token.LifeSec)
		if err != nil {
			log.Printf("ERROR: fail set expire for %s, error %s", "token:"+tokenIDStr, err.Error())
		}
	}
	_, err = conn.Do("HMSET", "access:"+tokenIDStr, "secret", token.Secret)
	if err != nil {
		log.Printf("ERROR: fail set key %s, val %s, error %s", "access:"+tokenIDStr, token.Secret, err.Error())
	} else {
		_, err = conn.Do("EXPIRE", "access:"+tokenIDStr, token.LifeSec)
		if err != nil {
			log.Printf("ERROR: fail set expire for %s, error %s", "access:"+tokenIDStr, err.Error())
		}
	}
}
