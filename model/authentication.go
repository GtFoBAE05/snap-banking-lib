package model

import "time"

type AccessToken struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
	ExpiresAt   time.Time
	Raw         interface{}
}
