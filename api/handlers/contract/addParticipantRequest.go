package contract

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type InviteClaims struct {
	Email      string `json:"email"`
	ContractID string `json:"contractId"`
	jwt.StandardClaims
}

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

		token, err := generateInviteToken(email, authExecutableCOntractKey)
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

		err = sendInviteEmail(email, inviteLink)
		if err != nil {
			errorhandler.ReturnError(c, err, "Failed to send invite email", http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invites sent successfully"})
}

func generateInviteToken(email, contractID string) (string, error) {
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

func sendInviteEmail(to, inviteLink string) error {
	from := os.Getenv("GOPROCESS_EMAIL")
	if from == "" {
		return fmt.Errorf("failed to find goprocess email")
	}

	password := os.Getenv("GOPROCESS_EMAIL_PASSWORD")
	if password == "" {
		return fmt.Errorf("failed to find goprocess email password")
	}

	fmt.Println("Sending email from: ", from)
	fmt.Println("Sending email password: ", password)

	msg := fmt.Sprintf("Subject: Contract Invitation\n\nPlease click the following link to accept the invitation to join the contract: %s", inviteLink)

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return fmt.Errorf("failed to find smtp host")
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		return fmt.Errorf("failed to find smtp port")
	}

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
