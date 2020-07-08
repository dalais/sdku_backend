package components

import (
	"encoding/json"
	"fmt"
)

// PostReqAnswer ...
type ReqAnswer struct {
	Error    int           `json:"error"`
	Data     interface{}   `json:"data"`
	Message  string        `json:"message"`
	ErrMesgs []interface{} `json:"err_mesgs"`
}

// IsEmptyData ...
func (pra *ReqAnswer) IsEmptyData() bool {
	data, err := json.Marshal(pra.Data)
	if err != nil {
		fmt.Println(err)
	}
	return string(data) == "{}"
}

// IsEmptyError ...
func (pra *ReqAnswer) IsEmptyError() bool {
	return pra.Error == 0
}
