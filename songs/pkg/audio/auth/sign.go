package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"time"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/golang-jwt/jwt/v5"
)

type Signer interface {
	Sign(claims Token) (string, error)
}

type signer struct {
	priv ed25519.PrivateKey
	exp  time.Duration
}

func NewSigner(encodedPriv string, exp time.Duration) (Signer, error) {
	decodedPriv, err := base64.StdEncoding.DecodeString(encodedPriv)
	if err != nil {
		return nil, e.From(err, fields.F("encodedPriv", encodedPriv))
	}

	priv := ed25519.PrivateKey(decodedPriv)

	return &signer{
		priv: priv,
		exp:  exp,
	}, nil
}

func (s signer) Sign(claims Token) (string, error) {
	if claims.Exp == 0 {
		claims.Exp = time.Now().Add(s.exp).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &claims)

	return token.SignedString(s.priv) //nolint:wrapcheck
}
