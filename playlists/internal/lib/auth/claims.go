package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Token struct {
	Subject  uuid.UUID `json:"sub"`
	IsArtist bool      `json:"is_artist"`
	Exp      int64     `json:"exp"`
}

func (t *Token) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

func (t *Token) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: time.Unix(t.Exp, 0)}, nil
}

func (t *Token) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

func (t *Token) GetIssuer() (string, error) {
	return "", nil
}

func (t *Token) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (t *Token) GetSubject() (string, error) {
	return t.Subject.String(), nil
}
