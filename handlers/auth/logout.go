package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dalais/sdku_backend/chttp"
	gl "github.com/dalais/sdku_backend/cmd/global"
)

// Logout ...
func Logout() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := gl.StoreSession.Get(r, "sessid")
		session.Options.MaxAge = -1
		err := session.Save(r, w)
		if err != nil {
			fmt.Printf("failed to delete session. Error: %v", err.Error())
		}
		sessionData := chttp.NewSessionData()
		sessionData.IsLogged = false
		sessionData.UserID = 0
		sessionData.Token = ""
		sessionData.Role = 0
		json.NewEncoder(w).Encode(sessionData)
	})
}
