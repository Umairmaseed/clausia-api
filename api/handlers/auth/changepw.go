package auth

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
)

type changePasswordForm struct {
	PreviousPassword string `json:"previousPassword" binding:"required"`
	ProposedPassword string `json:"proposedPassword" binding:"required"`
}

func (a *Auth) ChangePassword(c *gin.Context) {
	var form changePasswordForm
	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	accessToken, err := c.Cookie("accessToken")
	if err != nil {
		logger.Error(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	previousPassword := form.PreviousPassword
	proposedPassword := form.ProposedPassword

	input := cognito.ChangePasswordInput{
		AccessToken:      aws.String(accessToken),
		PreviousPassword: aws.String(previousPassword),
		ProposedPassword: aws.String(proposedPassword),
	}

	req, output := a.CognitoClient.ChangePasswordRequest(&input)

	err = req.Send()
	if err != nil {
		logger.Error(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, output.String())
}
