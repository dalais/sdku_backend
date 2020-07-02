package userstore

// User model struct
type User struct {
	ID            int64   `json:"id,omitempty"`
	Name          *string `json:"name,omitempty"`
	Email         string  `json:"email,omitempty" validate:"required,email,email_unique,real_email"`
	Password      string  `json:"password,omitempty" validate:"passwd"`
	Role          int     `json:"role"`
	EmailVerified *string `json:"email_verified,omitempty"`
	CrtdAt        *string `json:"crtd_at,omitempty"`
	ChngAt        *string `json:"chng_at,omitempty"`

	// For request
	Remember *bool `json:"remember,omitempty"`
}
