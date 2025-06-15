package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type TestAuthenticator struct{}

const secret = "test"

var testClaims = jwt.MapClaims{
	"sub": "65ea315e-ca1c-4af8-956b-57ed94378e94", //subject
	"exp": time.Now().Add(time.Hour).Unix(),
	"iss": "test-iss",
	"aud": "test-aud",
}

func (a *TestAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)

	tokenString, _ := token.SignedString([]byte(secret))

	return tokenString, nil
}

func (a *TestAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
