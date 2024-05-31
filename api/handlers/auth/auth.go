package auth

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/google/logger"
)

type Auth struct {
	CognitoClient   *cognito.CognitoIdentityProvider
	Region          string
	UserPoolID      string
	AppClientID     string
	AppClientSecret string
	CognitoURL      string
}

func NewAuth() Auth {

	conf := &aws.Config{
		Region:                        aws.String(os.Getenv("COGNITO_REGION")),
		CredentialsChainVerboseErrors: aws.Bool(true), // Enable verbose errors
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	auth := Auth{
		CognitoClient:   cognito.New(sess),
		Region:          os.Getenv("COGNITO_REGION"),
		UserPoolID:      os.Getenv("COGNITO_USER_POOL_ID"),
		AppClientID:     os.Getenv("COGNITO_APP_CLIENT_ID"),
		AppClientSecret: os.Getenv("COGNITO_APP_CLIENT_SECRET"),
	}

	auth.CognitoURL = "https://cognito-idp." + auth.Region + ".amazonaws.com/" + auth.UserPoolID + "/.well-known/jwks.json"

	return auth
}
