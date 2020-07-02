package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dalais/sdku_backend/cmd/cnf"
	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/store"
	userstore "github.com/dalais/sdku_backend/store/user"
	"golang.org/x/crypto/bcrypt"
)

// Login ...
func Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u *userstore.User
		var answer *components.PostReqAnswer
		answer = components.PostReqHandler(&u, w, r)
		user := userstore.User{
			Email:    u.Email,
			Password: u.Password,
			Role:     u.Role,
		}
		user, answer = LoginValidation(user, answer)
		if answer.Error == 0 {
			answer.Message = "Authentication is successful"
			token := components.GetToken(user)
			components.SendTokenToCookie(w, "access_token", token.Token, 1*time.Hour)
		}
		json.NewEncoder(w).Encode(answer)
	})
}

// LoginValidation ...
func LoginValidation(user userstore.User, answer *components.PostReqAnswer) (
	userstore.User, *components.PostReqAnswer) {
	if answer.Error == 0 {
		errMsgs := struct {
			Error string `json:"error"`
		}{}
		password := []byte(user.Password)
		answer = &components.PostReqAnswer{}
		row := store.Db.QueryRow(`SELECT id, email, role, password, crtd_at FROM users WHERE email=$1`, user.Email).Scan(
			&user.ID, &user.Email, &user.Role, &user.Password, &user.CrtdAt)
		if row != nil {
			errMsgs.Error = row.Error()
			answer.Data = nil
		}
		if row == nil {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), password)
			if err != nil {
				errMsgs.Error = err.Error()
			}
		}
		user.Password = ""
		if errMsgs.Error != "" {
			if cnf.Conf.DebugMode {
				answer.ErrMesgs = append(answer.ErrMesgs, errMsgs)
			}
			if !cnf.Conf.DebugMode {
				errMsgs.Error = "Incorrect email or password"
				answer.ErrMesgs = append(answer.ErrMesgs, errMsgs)
			}
			answer.Error = len(answer.ErrMesgs)
		}
		if len(answer.ErrMesgs) == 0 {
			answer.Data = user
		}
	}

	return user, answer
}
