package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/models"
	"github.com/go-playground/validator/v10"
)

// Register ...
func Register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u *models.User
		//var resp components.PostReqAnswer
		answer := components.PostReqHandler(&u, w, r)
		if answer.Error == 0 {
			answer = &components.PostReqAnswer{}
			v := validator.New()
			_ = v.RegisterValidation("passwd", func(fl validator.FieldLevel) bool {
				return len(fl.Field().String()) > 6
			})
			user := models.User{
				Email:    u.Email,
				Password: u.Password,
			}
			err := v.Struct(user)
			for _, e := range err.(validator.ValidationErrors) {
				msg := struct {
					Field string
					Msg   string
				}{
					Field: e.Field(),
					Msg:   e.Field(),
				}
				fmt.Println(msg)
				fmt.Println(e)
			}
		}
		json.NewEncoder(w).Encode(answer)
	})
}
