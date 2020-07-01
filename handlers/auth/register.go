package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/badoux/checkmail"
	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/store"
	userstore "github.com/dalais/sdku_backend/store/user"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// Registration ...
func Registration() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u *userstore.User
		var answer *components.PostReqAnswer
		answer = components.PostReqHandler(&u, w, r)
		user := userstore.User{
			Email:    u.Email,
			Password: u.Password,
		}
		answer = RegisterValidation(user, answer)
		if answer.Error == 0 {
			var newUser userstore.User
			components.Unmarshal(answer.Data, &newUser)

			password := []byte(newUser.Password)
			hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
			if err != nil {
				fmt.Println(err.Error())
				answer.Message = err.Error()
				return
			}
			sqlStatement := `
				INSERT INTO users (email, password, crtd_at)
					VALUES ($1, $2, $3)`
			_, err = store.Db.Exec(sqlStatement, newUser.Email, hashedPassword, time.Now())
			if err != nil {
				errM := struct {
					SQLError string `json:"sql_error"`
				}{
					SQLError: err.Error(),
				}
				answer.ErrMesgs = append(answer.ErrMesgs, errM)
				answer.Error = len(answer.ErrMesgs)
				panic(err)
			}

			user = userstore.User{}
			row := store.Db.QueryRow(`SELECT id, email FROM users WHERE email=$1`, newUser.Email).Scan(&user.ID, &user.Email)
			if row != nil {
				fmt.Println(row.Error())
			}
			if len(answer.ErrMesgs) == 0 {
				answer.Message = "User successfully created"
				u := struct {
					User userstore.User `json:"user"`
				}{
					User: user,
				}
				answer.Data = u
			}

		}
		json.NewEncoder(w).Encode(answer)
	})
}

// RegisterValidation ...
func RegisterValidation(model interface{}, answer *components.PostReqAnswer) *components.PostReqAnswer {
	if answer.Error == 0 {
		answer = &components.PostReqAnswer{}
		v := validator.New()
		_ = v.RegisterValidation("passwd", func(fl validator.FieldLevel) bool {
			return len(fl.Field().String()) > 6
		})
		_ = v.RegisterValidation("email_unique", func(fl validator.FieldLevel) bool {
			var email string
			row := store.Db.QueryRow("SELECT users.email FROM users WHERE email=$1", fl.Field().String())
			switch err := row.Scan(&email); err {
			case sql.ErrNoRows:
				return true
			case nil:
				return false
			default:
				fmt.Println(err)
				return false
			}
		})
		_ = v.RegisterValidation("real_email", func(fl validator.FieldLevel) bool {
			var err error
			err = checkmail.ValidateFormat(fl.Field().String())
			err = checkmail.ValidateHost(fl.Field().String())
			err = checkmail.ValidateHost(fl.Field().String())
			if smtpErr, ok := err.(checkmail.SmtpError); ok && err != nil {
				fmt.Printf("Code: %s, Msg: %s", smtpErr.Code(), smtpErr)
				return false
			}
			if err != nil {
				return false
			}
			return true
		})
		err := v.Struct(model)
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				var errMsg string
				if e.Tag() == "required" {
					errMsg = "This field is required"
				}
				if e.Tag() == "passwd" {
					errMsg = "The number of characters must be more than 6"
				}
				if e.Tag() == "email" {
					errMsg = "Invalid email address"
				}
				if e.Tag() == "email_unique" {
					errMsg = "This email is already in use"
				}
				if e.Tag() == "real_email" {
					errMsg = "Nonexistent email address"
				}
				msgData := struct {
					Field string `json:"field"`
					Msg   string `json:"message"`
				}{
					Field: e.Field(),
					Msg:   errMsg,
				}
				answer.ErrMesgs = append(answer.ErrMesgs, msgData)
			}
		}
		answer.Error = len(answer.ErrMesgs)
		answer.Data = model
	}

	return answer
}
