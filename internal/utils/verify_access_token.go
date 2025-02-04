package utils

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type CreateAccessTokenUtil struct{}

func NewCreateAccessTokenUtil() *CreateAccessTokenUtil {
	return &CreateAccessTokenUtil{}
}

func (b *CreateAccessTokenUtil) Validate(accessToken string) (*jwt.Token, jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET_JWT")), nil
	})
	return token, claims, err
}
