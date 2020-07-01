package auth

import (
	"encoding/json"
	"net/http"

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
		answer = LoginValidation(user, answer)
		/* if answer.Error == 0 {
			var newUser userstore.User
			components.Unmarshal(answer.Data, &newUser)
			if err != nil {
				fmt.Println(err.Error())
				answer.Message = err.Error()
				return
			}

		} */
		json.NewEncoder(w).Encode(answer)
	})
}

// LoginValidation ...
func LoginValidation(user userstore.User, answer *components.PostReqAnswer) *components.PostReqAnswer {
	if answer.Error == 0 {
		errMsgs := struct {
			Error string `json:"error"`
		}{}
		password := []byte(user.Password)
		answer = &components.PostReqAnswer{}
		row := store.Db.QueryRow(`SELECT id, email, role, password FROM users WHERE email=$1`, user.Email).Scan(
			&user.ID, &user.Email, &user.Role, &user.Password)
		if row != nil {
			errMsg := "Incorrect email or password"
			errMsgs.Error = errMsg
		}
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), password)
		if err != nil {
			errMsg := "Incorrect email or password"
			errMsgs.Error = errMsg
		}

		if errMsgs.Error != "" {
			answer.ErrMesgs = append(answer.ErrMesgs, errMsgs)
			answer.Error = len(answer.ErrMesgs)
		}
		if len(answer.ErrMesgs) == 0 {
			user.Password = ""
			answer.Data = user
		}
	}

	return answer
}
