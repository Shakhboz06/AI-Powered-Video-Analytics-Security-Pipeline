package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	Secret string
	Aud    string
	Iss    string
	Exp    time.Duration
}

type TokenClaims struct {
	jwt.RegisteredClaims
	Version int `json:"version"`
}

func NewJWTAuthenticator(secret, aud, iss string, exp time.Duration) *JWTAuthenticator {
	return &JWTAuthenticator{
		Secret: secret,
		Aud:    aud,
		Iss:    iss,
		Exp:    exp,
	}
}

func (tk *JWTAuthenticator) GenerateToken(userID int64, version int) (string, error) {

	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    tk.Iss,
			Audience:  jwt.ClaimStrings{tk.Aud},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tk.Exp)),
		},
		Version: version ,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(tk.Secret))

}

func (tk *JWTAuthenticator) ValidateToken(token string) (userID int64, tokenVer int, err error) {
	claims := &TokenClaims{}

	tokens, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok{
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(tk.Secret), nil
	},
	
	jwt.WithExpirationRequired(),
	jwt.WithAudience(tk.Aud),
	jwt.WithIssuer(tk.Iss),
	jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),	
	)
	if err != nil || !tokens.Valid{
		return 0, 0, errors.New("invalid or expired token")
	}

	userID, err = strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil{
		return 0, 0, errors.New("invalid user ID in token")
	}

	return userID, claims.Version, nil
}
