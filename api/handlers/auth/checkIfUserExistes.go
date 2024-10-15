package auth

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/google/logger"
)

func (a *Auth) CheckIfUserExistsAndGetEmail(username string) (bool, string, error) {
	conf := &aws.Config{
		Region:      aws.String(os.Getenv("COGNITO_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		logger.Error(err)
		return false, "", fmt.Errorf("failed to create AWS session: %v", err)
	}

	cognitoClient := cognito.New(sess)

	input := &cognito.AdminGetUserInput{
		UserPoolId: aws.String(a.UserPoolID),
		Username:   aws.String(username),
	}

	result, err := cognitoClient.AdminGetUser(input)
	if err != nil {
		if _, ok := err.(*cognito.ResourceNotFoundException); ok {
			return false, "", nil
		}
		return false, "", fmt.Errorf("error checking user existence in Cognito: %v", err)
	}

	var email string
	for _, attr := range result.UserAttributes {
		if *attr.Name == "email" {
			email = *attr.Value
			break
		}
	}

	return true, email, nil
}
