package components

// RAnswer ...
type RAnswer struct {
	Error    int      `json:"error"`
	Data     string   `json:"data"`
	Message  string   `json:"message"`
	ErrMesgs []string `json:"err_mesgs"`
}
