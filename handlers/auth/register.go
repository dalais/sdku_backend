package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/store"
	userstore "github.com/dalais/sdku_backend/store/user"
	"github.com/go-playground/validator/v10"
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
		answer = Validation(user, answer)
		if answer.Error == 0 {
			var newUser userstore.User
			components.Unmarshal(answer.Data, &newUser)
			fmt.Println(newUser)
			/* sqlStatement := `
				INSERT INTO users (email, password, crtd_at, chng_at)
					VALUES ($1, $2, $3, $4)`
			_, err = store.Db.Exec(sqlStatement, )
			if err != nil {
				panic(err)
			} */
		}
		json.NewEncoder(w).Encode(answer)
	})
}

// Validation ...
func Validation(model interface{}, answer *components.PostReqAnswer) *components.PostReqAnswer {
	if answer.Error == 0 {
		answer = &components.PostReqAnswer{}
		v := validator.New()
		_ = v.RegisterValidation("passwd", func(fl validator.FieldLevel) bool {
			return len(fl.Field().String()) > 6
		})
		_ = v.RegisterValidation("email_unique", func(fl validator.FieldLevel) bool {
			var email string
			var id int
			row := store.Db.QueryRow(`SELECT * FROM users WHERE email=$1`, fl.Field().String())
			switch err := row.Scan(&id, &email); err {
			case sql.ErrNoRows:
				return true
			case nil:
				return false
			default:
				panic(err)
			}
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
					errMsg = "Field validation for 'Email' failed"
				}
				if e.Tag() == "email_unique" {
					errMsg = "This email is already in use"
				}
				msg := struct {
					Field string `json:"field"`
					Msg   string `json:"message"`
				}{
					Field: e.Field(),
					Msg:   errMsg,
				}
				msgData, err := json.Marshal(msg)
				if err != nil {
					log.Fatal(err)
				}
				answer.ErrMesgs = append(answer.ErrMesgs, string(msgData))
			}
		}
		answer.Error = len(answer.ErrMesgs)
		answer.Data = model
	}

	return answer
}
