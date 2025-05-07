package auth

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/utils"

	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type VerifyAccountForm struct {
	UserName string `json:"username" binding:"required"`
	OTP      string `json:"otp" binding:"required"`
}

func (a *Auth) VerifyAccount(c *gin.Context) {
	var form VerifyAccountForm
	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	username := form.UserName
	otp := form.OTP

	var secretHash string

	if a.AppClientSecret != "" {
		secretHash = utils.ComputeSecretHash(a.AppClientSecret, username, a.AppClientID)
	}

	user := &cognito.ConfirmSignUpInput{
		SecretHash:       &secretHash,
		ConfirmationCode: aws.String(otp),
		Username:         aws.String(username),
		ClientId:         aws.String(a.AppClientID),
	}

	_, err := a.CognitoClient.ConfirmSignUp(user)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
