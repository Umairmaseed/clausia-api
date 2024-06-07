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

type signInForm struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"newPassword"`
}

const USER_PASS_FLOW = "USER_PASSWORD_AUTH"

func (a *Auth) SignIn(c *gin.Context) {
	var form signInForm
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

	var IdToken *string
	var AccessToken *string
	var RefreshToken *string

	res, err := a.CognitoClient.InitiateAuth(authTry)
	if err != nil {
		if strings.Contains(err.Error(), "UserNotConfirmedException") {
			getUserInput := &cognito.AdminGetUserInput{
				Username:   aws.String(username),
				UserPoolId: &a.UserPoolID,
			}
			user, err := a.CognitoClient.AdminGetUser(getUserInput)
			if err != nil {
				logger.Error(err)
				c.JSON(http.StatusBadRequest, err.Error())
				return
			}

			email := ""
			for _, p := range user.UserAttributes {
				if *p.Name == "email" {
					email = *p.Value
				}
			}

			resMap := map[string]interface{}{
				"message":  "UserNotConfirmedException",
				"email":    email,
				"username": username,
			}

			c.JSON(http.StatusOK, resMap)
			return
		}
		logger.Error(err)

		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if res.ChallengeName != nil && *res.ChallengeName == cognito.ChallengeNameTypeNewPasswordRequired {
		input := cognito.AdminRespondToAuthChallengeInput{
			ChallengeName: res.ChallengeName,
			ClientId:      &a.AppClientID,
			UserPoolId:    &a.UserPoolID,
			Session:       res.Session,
			ChallengeResponses: map[string]*string{
				"NEW_PASSWORD": aws.String(form.NewPassword),
				"USERNAME":     aws.String(form.Username),
				"SECRET_HASH":  aws.String(secretHash),
			},
		}

		req, output := a.CognitoClient.AdminRespondToAuthChallengeRequest(&input)

		err := req.Send()
		if err != nil {
			logger.Error(err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		IdToken = output.AuthenticationResult.IdToken
		AccessToken = output.AuthenticationResult.AccessToken
		RefreshToken = output.AuthenticationResult.RefreshToken
	} else {
		IdToken = res.AuthenticationResult.IdToken
		AccessToken = res.AuthenticationResult.AccessToken
		RefreshToken = res.AuthenticationResult.RefreshToken
	}

	secure := false
	c.SetSameSite(http.SameSiteDefaultMode)

	c.SetCookie("idToken", *IdToken, 86400, "", "/", secure, true)
	c.SetCookie("accessToken", *AccessToken, 86400, "", "/", secure, true)
	c.SetCookie("refreshToken", *RefreshToken, 86400, "", "/", secure, true)

	c.Status(http.StatusOK)
}
