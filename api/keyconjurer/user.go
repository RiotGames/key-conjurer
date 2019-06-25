package keyconjurer

// User represents a user
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewUser creates a User with the provided info
func NewUser(username, password string) *User {
	return &User{
		Username: username,
		Password: password}
}
