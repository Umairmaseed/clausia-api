package auth

import (
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/utils"
)

type confirmForgotPasswordForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	OTP      string `json:"otp" binding:"required"`
}

func (a *Auth) ConfirmForgotPassword(c *gin.Context) {
	var form confirmForgotPasswordForm
	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	username := form.Username
	password := form.Password
	otp := form.OTP
	clientID := os.Getenv("COGNITO_APP_CLIENT_ID")
	clientSecret := os.Getenv("COGNITO_APP_CLIENT_SECRET")

	input := cognito.ConfirmForgotPasswordInput{
		ClientId:         aws.String(clientID),
		Username:         aws.String(username),
		Password:         aws.String(password),
		SecretHash:       aws.String(utils.ComputeSecretHash(clientSecret, username, clientID)),
		ConfirmationCode: aws.String(otp),
	}

	req, output := a.CognitoClient.ConfirmForgotPasswordRequest(&input)

	err := req.Send()
	if err != nil {
		logger.Error(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, output.String())
}
