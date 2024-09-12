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

type addParticipantsRequestForm struct {
	AutoExecutableContract map[string]interface{}   `form:"autoExecutableContract" binding:"required"`
	Participants           []map[string]interface{} `form:"participants" binding:"required"`
}

func AddParticipantRequest(c *gin.Context) {
	var form addParticipantsRequestForm
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	authExecutableCOntractKey, ok := form.AutoExecutableContract["@key"].(string)
	if !ok {
		errorhandler.ReturnError(c, fmt.Errorf("invalid key for auto executable contract"), "Invalid key", http.StatusBadRequest)
		return
	}

	for _, participant := range form.Participants {
		ledgerKey, ok := participant["@key"].(string)
		if !ok {
			errorhandler.ReturnError(c, fmt.Errorf("invalid key for participant"), "Invalid key", http.StatusBadRequest)
			return
		}

		signerAsset, err := chaincode.GetSigner(ledgerKey)
		if err != nil {
			errorhandler.ReturnError(c, err, "Failed to retrieve signer asset", http.StatusInternalServerError)
			return
		}

		email, ok := signerAsset["email"].(string)
		if !ok || email == "" {
			errorhandler.ReturnError(c, fmt.Errorf("signer asset does not contain a valid email"), "Invalid email in signer asset", http.StatusInternalServerError)
			return
		}

		token, err := utils.GenerateInviteToken(email, authExecutableCOntractKey, jwtSecret)
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

		msg := fmt.Sprintf("Subject: Contract Invitation\n\nPlease click the following link to accept the invitation to join the contract: %s", inviteLink)

		err = utils.SendInviteEmail(email, msg)
		if err != nil {
			errorhandler.ReturnError(c, err, "Failed to send invite email", http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invites sent successfully"})
}
