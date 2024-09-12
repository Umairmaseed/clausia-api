package contract

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

type shareTemplateRequestForm struct {
	Template map[string]interface{}   `form:"template" binding:"required"`
	Users    []map[string]interface{} `form:"users" binding:"required"`
}

func ShareTemplate(c *gin.Context) {
	var form shareTemplateRequestForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		errorhandler.ReturnError(c, fmt.Errorf("email not found in headers"), "email not found in headers", http.StatusBadRequest)
		return
	}

	userKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to find user key", http.StatusInternalServerError)
		return
	}

	contractAsset, err := chaincode.SearchAsset(form.Template)
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

	templateOwner, ok := firstResult["creator"].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("could not find creator of th template"), "could not find creator of th template", http.StatusInternalServerError)
		return
	}

	isPublic, ok := firstResult["public"].(bool)
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("could not find public field of th template"), "could not find public field of th template", http.StatusInternalServerError)
		return
	}

	if templateOwner["@key"] != userKey && !isPublic {
		errorhandler.ReturnError(c, fmt.Errorf("only the creator of the contract can send invite"), "only the owner of the contract can send invite", http.StatusBadRequest)
		return
	}

	templateKey, ok := form.Template["@key"].(string)
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid key for template"), "Invalid key", http.StatusBadRequest)
		return
	}

	for _, user := range form.Users {
		ledgerKey, ok := user["@key"].(string)
		if !ok {
			errorhandler.ReturnError(c, fmt.Errorf("invalid key for user"), "Invalid key", http.StatusBadRequest)
			return
		}

		signerAsset, err := chaincode.GetSigner(ledgerKey)
		if err != nil {
			errorhandler.ReturnError(c, err, "Failed to retrieve user asset", http.StatusInternalServerError)
			return
		}

		email, ok := signerAsset["email"].(string)
		if !ok || email == "" {
			errorhandler.ReturnError(c, fmt.Errorf("user asset does not contain a valid email"), "Invalid email in user asset", http.StatusInternalServerError)
			return
		}

		token, err := utils.GenerateInviteToken(email, templateKey, jwtSecret)
		if err != nil {
			errorhandler.ReturnError(c, err, "Failed to generate invite token", http.StatusInternalServerError)
			return
		}

		inviteLinkBase := os.Getenv("INVITE_LINK")
		if inviteLinkBase == "" {
			errorhandler.ReturnError(c, nil, "Failed to find invite link", http.StatusInternalServerError)
			return
		}

		inviteLink := inviteLinkBase + token

		msg := fmt.Sprintf("Subject: Template Invitation\n\nYou have been invited to view a template.\nPlease click the following link to view the invitation: %s", inviteLink)

		err = utils.SendInviteEmail(email, msg)
		if err != nil {
			errorhandler.ReturnError(c, err, "Failed to send invite email", http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invites sent successfully"})
}
