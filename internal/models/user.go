package models

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID       int64
	Login    string
	PassHash []byte
}

type UserClaims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}
