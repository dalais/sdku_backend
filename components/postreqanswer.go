package components

import (
	"encoding/json"
	"fmt"
)

// PostReqAnswer ...
type PostReqAnswer struct {
	Error    int         `json:"error"`
	Data     interface{} `json:"data"`
	Message  string      `json:"message"`
	ErrMesgs []string    `json:"err_mesgs"`
}

// IsEmptyData ...
func (pra *PostReqAnswer) IsEmptyData() bool {
	data, err := json.Marshal(pra.Data)
	if err != nil {
		fmt.Println(err)
	}
	return string(data) == "{}"
}

// IsEmptyError ...
func (pra *PostReqAnswer) IsEmptyError() bool {
	return pra.Error == 0
}
