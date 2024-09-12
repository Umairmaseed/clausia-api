package utils

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateInviteToken(email, contractID string, jwtSecret []byte) (string, error) {
	expiryDurationStr := os.Getenv("INVITE_EXPiRY_TIME")

	expiryDuration, err := strconv.Atoi(expiryDurationStr)
	if err != nil {
		return "", fmt.Errorf("error converting expiry duration to int: %w", err)
	}

	expirationTime := time.Now().Add(time.Duration(expiryDuration) * time.Hour)

	claims := &InviteClaims{
		Email:      email,
		ContractID: contractID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
