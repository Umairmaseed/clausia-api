package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

func ViewSharedTemplate(c *gin.Context) {

	tokenString := c.Query("token")
	if tokenString == "" {
		errorhandler.ReturnError(c, fmt.Errorf("token is required"), "Token is required", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		errorhandler.ReturnError(c, fmt.Errorf("email not found in headers"), "email not found in headers", http.StatusBadRequest)
		return
	}

	claims, err := utils.VerifyInviteToken(tokenString, jwtSecret)
	if err != nil {
		errorhandler.ReturnError(c, err, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	UserEmail := claims.Email
	templateID := claims.ContractID

	template := map[string]interface{}{
		"@key":       templateID,
		"@assetType": "template",
	}

	if UserEmail != email {
		errorhandler.ReturnError(c, fmt.Errorf("user email does not match token email"), "user email does not match token email", http.StatusBadRequest)
		return
	}

	contractAsset, err := chaincode.SearchAsset(template)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find template asset", http.StatusInternalServerError)
		return
	}

	results, ok := contractAsset["result"].([]interface{})
	if !ok || len(results) == 0 {
		errorhandler.ReturnError(c, fmt.Errorf("no results found in template asset"), "no results found in template asset", http.StatusInternalServerError)
		return
	}

	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid result format"), "invalid result format", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Template retrieved successfully",
		"template": firstResult,
	})
}
