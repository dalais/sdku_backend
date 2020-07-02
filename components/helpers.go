package components

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/dalais/sdku_backend/cmd/cnf"
)

// Unmarshal ...
func Unmarshal(data interface{}, relate interface{}) {
	jsdata, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	json.Unmarshal([]byte(string(jsdata)), &relate)
}

// RandomString generator
func RandomString(n int) string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// HandleAnswerError ...
func HandleAnswerError(err error, answer *PostReqAnswer, msg string) {
	if err != nil {
		errM := struct {
			Error string `json:"error"`
		}{}
		if cnf.Conf.DebugMode {
			errM.Error = err.Error()
		}
		if !cnf.Conf.DebugMode {
			errM.Error = msg
		}

		answer.ErrMesgs = append(answer.ErrMesgs, errM)
		answer.Error = len(answer.ErrMesgs)
		fmt.Println(err)
	}
}
