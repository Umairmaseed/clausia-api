package utils

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

type InviteClaims struct {
	Email      string `json:"email"`
	ContractID string `json:"contractId"`
	jwt.StandardClaims
}

func VerifyInviteToken(tokenString string, jwtSecret []byte) (*InviteClaims, error) {
	claims := &InviteClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
