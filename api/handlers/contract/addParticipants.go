package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

func AddParticipants(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		errorhandler.ReturnError(c, fmt.Errorf("token is required"), "Token is required", http.StatusBadRequest)
		return
	}

	claims, err := verifyInviteToken(tokenString)
	if err != nil {
		errorhandler.ReturnError(c, err, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	participantEmail := claims.Email
	contractID := claims.ContractID

	autoExecutableContract := map[string]interface{}{
		"@key":       contractID,
		"@assetType": "autoExecutableContract",
	}

	participantKey, err := utils.SearchAndReturnSignerKey(participantEmail)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find user key", http.StatusInternalServerError)
		return
	}

	contractAsset, err := chaincode.SearchAsset(autoExecutableContract)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find contract asset", http.StatusInternalServerError)
		return
	}

	results, ok := contractAsset["result"].([]interface{})
	if !ok || len(results) == 0 {
		errorhandler.ReturnError(c, fmt.Errorf("no results found in contract asset"), "no results found in contract asset", http.StatusInternalServerError)
		return
	}

	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid result format"), "invalid result format", http.StatusInternalServerError)
		return
	}

	participants := []map[string]interface{}{
		{
			"@key":       participantKey,
			"@assetType": "user",
		},
	}

	reqMap := map[string]interface{}{
		"autoExecutableContract": firstResult,
		"participants":           participants,
	}

	updatedContractAsset, err := chaincode.AddParticipants(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add participants to contract", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"contract": updatedContractAsset})
}

func verifyInviteToken(tokenString string) (*InviteClaims, error) {
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
