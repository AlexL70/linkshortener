package models

// User represents an authenticated application user at the business layer.
type User struct {
	ID       int64
	UserName string
}
