package auth

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/certs"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"

	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type signUpForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

func (a *Auth) SignUp(c *gin.Context) {
	var form signUpForm
	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	username := form.Username
	password := form.Password
	email := form.Email

	var secretHash string

	if a.AppClientSecret != "" {
		secretHash = utils.ComputeSecretHash(a.AppClientSecret, username, a.AppClientID)
	}

	user := &cognito.SignUpInput{
		Username:   aws.String(username),
		Password:   aws.String(password),
		SecretHash: &secretHash,
		ClientId:   aws.String(a.AppClientID),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("name"),
				Value: aws.String(username),
			},
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
	}

	out, err := a.CognitoClient.SignUp(user)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	cert, err := certs.CreateIdentityHandler(c, username, form.Name, password)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	certName := username + "_cert.pfx"
	_, err = utils.UploadCertToS3(cert, certName)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	sub := *out.UserSub

	c.JSON(http.StatusOK, sub)
}
