package common

import "github.com/golang-jwt/jwt/v5"

// User represents a registered user
type User struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Hash     []byte `json:"hash"`
}

// Credentials represents the credential DTO holder
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims represents the claims DTO holder
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
