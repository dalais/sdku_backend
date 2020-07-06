package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dalais/sdku_backend/cmd/cnf"
	"github.com/dalais/sdku_backend/components"
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
		var answer *components.PostReqAnswer
		// Handle http request
		answer = components.PostReqHandler(&u, w, r)

		user := userstore.User{
			Email:    u.Email,
			Password: u.Password,
			Role:     u.Role,
			Remember: u.Remember,
		}

		// Validation fields
		user, answer = loginValidation(user, answer)

		var token components.TokenObj

		// Create and store token
		if answer.IsEmptyError() {
			token = storeToken(user, answer)
		}

		// Sessions
		session, _ := cnf.StoreSession.Get(r, "sessid")
		session.Values["rmb"] = user.Remember

		var tm int
		if *user.Remember == true {
			tm = 60 //86400 * 7
		} else {
			tm = -1
		}
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   tm,
			HttpOnly: true,
		}
		err := session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If everything is in order we set a notification and send the token to cookies
		if answer.IsEmptyError() {
			answer.Message = "Authentication is successful"
			components.SendTokenToCookie(w, "access_token", token.Token, 1*time.Hour)
		}

		json.NewEncoder(w).Encode(answer)
	})
}

// LoginValidation ...
func loginValidation(user userstore.User, answer *components.PostReqAnswer) (
	userstore.User, *components.PostReqAnswer) {
	if answer.Error == 0 {
		// new error struct for answer.ErrMesgs
		errMsgs := struct {
			Error string `json:"error"`
		}{}
		// New answer struct
		answer = &components.PostReqAnswer{}

		// Getting the password sent in the request
		password := []byte(user.Password)

		// Select user from db
		row := store.Db.QueryRow(`SELECT id, email, role, password, crtd_at FROM users WHERE email=$1`, user.Email).Scan(
			&user.ID, &user.Email, &user.Role, &user.Password, &user.CrtdAt)
		if row != nil {
			components.HandleAnswerError(row, answer, custErrMsg)
			answer.Data = nil
		}

		// If record exist, compare passwords
		if row == nil {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), password)
			components.HandleAnswerError(err, answer, custErrMsg)
		}
		user.Password = ""

		// Errors handling
		if errMsgs.Error != "" {
			components.HandleAnswerError(&LoginError{}, answer, custErrMsg)
		}

		// If there are no errors, we set user data for the answer.Data field
		if len(answer.ErrMesgs) == 0 {
			answer.Data = user
		}
	}

	return user, answer
}

// TODO ...
func storeToken(user userstore.User, answer *components.PostReqAnswer) components.TokenObj {

	var tokenID int64
	// Create sercret string
	secret := components.RandomString(32)

	// BEGIN transaction
	tx, err := store.Db.Begin()
	components.HandleAnswerError(err, answer, custErrMsg)

	createTokensql := `
				INSERT INTO auth_tokens (user_id)
					VALUES ($1) RETURNING id`
	err = tx.QueryRow(createTokensql, user.ID).Scan(&tokenID)
	if err != nil {
		tx.Rollback()
		components.HandleAnswerError(err, answer, custErrMsg)
	}

	// insert record into table2, referencing the first record from table1
	_, err = tx.Exec("INSERT INTO auth_access(token_id, secret) VALUES($1, $2)", tokenID, secret)
	if err != nil {
		tx.Rollback()
		components.HandleAnswerError(err, answer, custErrMsg)
	}
	// Create token object
	token := components.GetToken(secret, tokenID)

	// insert access_token
	_, err = tx.Exec("UPDATE auth_tokens SET access_token=$1 WHERE id=$2", token.Token, tokenID)
	if err != nil {
		tx.Rollback()
		components.HandleAnswerError(err, answer, custErrMsg)
	}

	// COMMIT the transaction
	components.HandleAnswerError(tx.Commit(), answer, custErrMsg)
	if !answer.IsEmptyError() {
		token.Token = ""
		answer.Data = nil
	}

	return token
}
