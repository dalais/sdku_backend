package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

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

		// Handling post request
		answer = components.PostReqHandler(&u, w, r)
		user := userstore.User{
			Email:    u.Email,
			Password: u.Password,
		}

		// Validation
		answer = registerValidation(user, answer)
		if answer.Error == 0 {
			// new struct for new user
			var newUser userstore.User
			// Set query data for fields the newUser
			components.Unmarshal(answer.Data, &newUser)

			// Creat hash from password
			password := []byte(newUser.Password)
			hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
			if err != nil {
				fmt.Println(err.Error())
				answer.Message = err.Error()
				return
			}

			// Storing data
			sqlStatement := `
				INSERT INTO users (email, password)
					VALUES ($1, $2) RETURNING id, email, crtd_at`
			err = store.Db.QueryRow(sqlStatement, newUser.Email, hashedPassword).Scan(
				&user.ID, &user.Email, &user.CrtdAt)
			user.Password = ""
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

// registerValidation ...
func registerValidation(model interface{}, answer *components.PostReqAnswer) *components.PostReqAnswer {
	if answer.Error == 0 {
		// Set empty answer struct
		answer = &components.PostReqAnswer{}

		// Initialization new validation instance
		v := validator.New()

		// Custom validation for password length
		_ = v.RegisterValidation("passwd", func(fl validator.FieldLevel) bool {
			return len(fl.Field().String()) > 6
		})
		// Checking existing email in database
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
		// Checking existing email. Is used https://github.com/badoux/checkmail
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

		// Error handling
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
