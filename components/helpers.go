package components

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	gl "github.com/dalais/sdku_backend/cmd/global"
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

// HandleAnswerError ...
func HandleAnswerError(err error, answer *ReqAnswer, msg string) {
	if err != nil {
		errM := struct {
			Error string `json:"error"`
		}{}
		if gl.Conf.DebugMode {
			errM.Error = err.Error()
			fmt.Println(err)
		}
		if !gl.Conf.DebugMode {
			errM.Error = msg
		}

		answer.ErrMesgs = append(answer.ErrMesgs, errM)
		answer.Error = len(answer.ErrMesgs)
	}
}

// RandomString ...
func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
