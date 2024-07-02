package auth

import (
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/google/logger"
)

type verifyEmail struct {
	Code string `form:"code" binding:"required"`
}

func (a *Auth) ConfirmNewEmail(c *gin.Context) {
	var form verifyEmail

	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind form data", http.StatusBadRequest)
		return
	}

	cookie, err := c.Cookie("accessToken")
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to get access token from cookie", http.StatusUnauthorized)
		return
	}

	conf := &aws.Config{
		Region:      aws.String(os.Getenv("COGNITO_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		logger.Error(err)
		errorhandler.ReturnError(c, err, "Failed to create AWS session", http.StatusInternalServerError)
		return
	}

	cognitoClient := cognito.New(sess)

	input := &cognito.VerifyUserAttributeInput{
		AccessToken:   aws.String(cookie),
		AttributeName: aws.String("email"),
		Code:          aws.String(form.Code),
	}

	_, err = cognitoClient.VerifyUserAttribute(input)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to verify email in Cognito", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}
