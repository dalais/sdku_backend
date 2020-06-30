package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dalais/sdku_backend/components"
	userstore "github.com/dalais/sdku_backend/store/user"
	"github.com/go-playground/validator/v10"
)

// Register ...
func Register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u *userstore.User
		var answer *components.PostReqAnswer
		answer = components.PostReqHandler(&u, w, r)
		user := userstore.User{
			Email:    u.Email,
			Password: u.Password,
		}
		answer = Validation(user, answer)
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
