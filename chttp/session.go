package chttp

// SessionData ...
type SessionData struct {
	IsLogged bool   `json:"is_logged"`
	UserID   int64  `json:"user_id"`
	Role     int    `json:"role"`
	Token    string `json:"token"`
	CSRF     string `json:"csrf"`
}

// NewSessionData ...
func NewSessionData() SessionData {
	session := SessionData{}
	session.IsLogged = false
	session.UserID = 0
	session.Role = 0
	session.Token = ""
	session.CSRF = ""
	return session
}
