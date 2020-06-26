package components

// PostReqAnswer ...
type PostReqAnswer struct {
	Error    int         `json:"error"`
	Data     interface{} `json:"data"`
	Message  string      `json:"message"`
	ErrMesgs []string    `json:"err_mesgs"`
}

// IsEmptyData ...
func (pra *PostReqAnswer) IsEmptyData() bool {
	return pra.Data == nil
}
