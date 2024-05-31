package auth

import (
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type forgotPasswordForm struct {
	Username string `json:"username" binding:"required"`
}

func (a *Auth) ForgotPassword(c *gin.Context) {
	var form forgotPasswordForm
	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	username := form.Username
	clientID := os.Getenv("COGNITO_APP_CLIENT_ID")
	clientSecret := os.Getenv("COGNITO_APP_CLIENT_SECRET")

	input := cognito.ForgotPasswordInput{
		ClientId:   aws.String(clientID),
		Username:   aws.String(username),
		SecretHash: aws.String(utils.ComputeSecretHash(clientSecret, username, clientID)),
	}

	req, output := a.CognitoClient.ForgotPasswordRequest(&input)

	err := req.Send()
	if err != nil {
		logger.Error(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, output.String())
}
