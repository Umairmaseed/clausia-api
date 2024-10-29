package auth

import (
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type passwordCheckForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (a *Auth) CheckPw(c *gin.Context) {
	var form passwordCheckForm
	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	username := form.Username
	password := form.Password

	params := map[string]*string{
		"USERNAME": aws.String(username),
		"PASSWORD": aws.String(password),
	}

	var secretHash string
	if a.AppClientSecret != "" {
		secretHash = utils.ComputeSecretHash(a.AppClientSecret, username, a.AppClientID)
		params["SECRET_HASH"] = aws.String(secretHash)
	}

	authTry := &cognito.InitiateAuthInput{
		AuthFlow:       aws.String(USER_PASS_FLOW),
		AuthParameters: params,
		ClientId:       aws.String(a.AppClientID),
	}

	_, err := a.CognitoClient.InitiateAuth(authTry)
	if err != nil {
		if strings.Contains(err.Error(), "NotAuthorizedException") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Incorrect username or password"})
			return
		}
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password is correct"})
}
