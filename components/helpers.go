package components

import (
	"encoding/json"
	"fmt"
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
