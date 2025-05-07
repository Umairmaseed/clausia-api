package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/utils"
)

type updateSigner struct {
	Email string `form:"email"`
	Phone string `form:"phone"`
}

func (a *Auth) UpdateEmailOrPhone(c *gin.Context) {
	var form updateSigner

	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind form data", http.StatusBadRequest)
		return
	}

	email := c.Request.Header.Get("Email")
	if email == "" {
		errorhandler.ReturnError(c, fmt.Errorf("email not found in headers"), "email not found in headers", http.StatusBadRequest)
		return
	}

	username := c.Request.Header.Get("Username")
	if username == "" {
		errorhandler.ReturnError(c, fmt.Errorf("username not found in headers"), "username not found in headers", http.StatusBadRequest)
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		errorhandler.ReturnError(c, err, err.Error(), http.StatusInternalServerError)
		return
	}

	conf := &aws.Config{
		Region:      aws.String(os.Getenv("COGNITO_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	cognitoClient := cognito.New(sess)

	input := &cognito.AdminUpdateUserAttributesInput{
		UserPoolId: aws.String(a.UserPoolID),
		Username:   aws.String(username),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(form.Email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("false"),
			},
		},
	}

	_, err = cognitoClient.AdminUpdateUserAttributes(input)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to update user email in Cognito", http.StatusInternalServerError)
		return
	}

	updatesMap := map[string]interface{}{}
	if form.Email != "" {
		updatesMap["email"] = form.Email
	}
	if form.Phone != "" {
		updatesMap["phone"] = form.Phone
	}

	signerMap := map[string]interface{}{
		"@assetType": "user",
		"@key":       signerKey,
	}

	_, err = chaincode.UpdateSigner(signerMap, updatesMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to update signer in the blockchain", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User email and/or phone updated successfully",
	})
}
