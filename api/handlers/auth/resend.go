package auth

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/utils"
)

type resendCodeForm struct {
	Username string `json:"username" binding:"required"`
}

func (a *Auth) ResendCode(c *gin.Context) {
	var form resendCodeForm
	var err error

	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	username := form.Username

	var secretHash string
	if a.AppClientSecret != "" {
		secretHash = utils.ComputeSecretHash(a.AppClientSecret, username, a.AppClientID)
	}

	params := &cognito.ResendConfirmationCodeInput{
		SecretHash: &secretHash,
		Username:   aws.String(username),
		ClientId:   aws.String(a.AppClientID),
	}

	req, out := a.CognitoClient.ResendConfirmationCodeRequest(params)

	err = req.Send()
	if err != nil {
		logger.Error(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, out.String())
}
