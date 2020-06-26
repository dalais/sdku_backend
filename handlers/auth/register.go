package auth

import (
	"encoding/json"
	"net/http"

	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/models"
)

// Register ...
func Register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u *models.User
		answer := components.PostReqHandler(&u, w, r)
		json.NewEncoder(w).Encode(answer)
	})
}
