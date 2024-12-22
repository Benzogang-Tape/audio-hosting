package auth

import (
	"crypto/ed25519"
	"encoding/base64"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/golang-jwt/jwt/v5"
)

var ErrParsing = e.New("token parsing failed")

type Parser interface {
	Parse(token string) (Token, error)
}

type parser struct {
	pub ed25519.PublicKey
}

func NewParser(encodedPub string) (Parser, error) {
	decodedPub, err := base64.StdEncoding.DecodeString(encodedPub)
	if err != nil {
		return nil, e.From(err, fields.F("encodedPub", encodedPub))
	}

	return &parser{
		pub: ed25519.PublicKey(decodedPub),
	}, nil
}

func (p parser) Parse(token string) (Token, error) {
	var claims Token

	_, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (any, error) {
		return p.pub, nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
	)
	if err != nil {
		return Token{}, ErrParsing.Wrap(err)
	}

	return claims, nil
}
