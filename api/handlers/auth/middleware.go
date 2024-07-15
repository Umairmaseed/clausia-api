package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
	"github.com/joho/godotenv"
)

type requestClaims struct {
	Username         string `json:"cognito:username"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"email_verified"`
	Name             string `json:"name"`
	CustomChaincodes string `json:"custom:chaincodes"`
	jwt.StandardClaims
}

func (a *Auth) AuthMiddleware() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		godotenv.Load(".env")

		tokenString, err := c.Cookie("idToken")
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		tokenExpired := false

		token, err := jwt.ParseWithClaims(tokenString, &requestClaims{}, a.checkToken)
		if err != nil && !strings.Contains(err.Error(), "token is expired") {
			logger.Error(err.Error())
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		} else if err != nil && strings.Contains(err.Error(), "token is expired") {
			tokenExpired = true
		}

		claims, _ := token.Claims.(*requestClaims)
		if !token.Valid || tokenExpired {
			output, err := a.refreshAuth(c, claims)
			if err != nil {
				logger.Error(fmt.Errorf("invalid token"))
				c.JSON(http.StatusUnauthorized, fmt.Errorf("invalid token"))
				c.Abort()
				return
			}

			IdToken := output.AuthenticationResult.IdToken
			AccessToken := output.AuthenticationResult.AccessToken

			t, err := jwt.ParseWithClaims(*IdToken, &requestClaims{}, a.checkToken)
			if err != nil {
				logger.Error(err.Error())
				c.JSON(http.StatusUnauthorized, err.Error())
				c.Abort()
				return
			}

			if !t.Valid {
				logger.Error(err.Error())
				c.JSON(http.StatusUnauthorized, err.Error())
				c.Abort()
				return
			}

			c.SetSameSite(http.SameSiteLaxMode)

			c.SetCookie("idToken", *IdToken, 86400, "", c.Request.Host, false, true)
			c.SetCookie("accessToken", *AccessToken, 86400, "", c.Request.Host, false, true)
		}

		idClaimsJSON, _ := json.Marshal(claims)

		// Avoid header injection
		c.Request.Header.Del("UserId")
		c.Request.Header.Del("Username")
		c.Request.Header.Del("email")
		c.Request.Header.Del("emailverified")

		// Add headers
		username := claims.Username
		c.Writer.Header().Set("Username", username)
		c.Request.Header.Add("Username", username)
		c.Request.Header.Add("UserId", claims.Subject)
		c.Request.Header.Add("email", claims.Email)
		c.Request.Header.Add("emailverified", fmt.Sprintf("%t", claims.EmailVerified))
		c.Request.Header.Add("Idclaims", string(idClaimsJSON))

		c.Next()
	}
	return fn
}

func (a *Auth) refreshAuth(c *gin.Context, claims *requestClaims) (*cognitoidentityprovider.AdminInitiateAuthOutput, error) {
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		return nil, err
	}

	secretHash := utils.ComputeSecretHash(a.AppClientSecret, claims.Username, a.AppClientID)

	input := cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": &refreshToken,
			"SECRET_HASH":   &secretHash,
		},
		ClientId:   &a.AppClientID,
		UserPoolId: &a.UserPoolID,
	}

	o, err := a.CognitoClient.AdminInitiateAuth(&input)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (a *Auth) checkToken(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	_, ok := token.Claims.(*requestClaims)
	if !ok {
		return nil, fmt.Errorf("request claims not ok")
	}

	kid := token.Header["kid"]

	keys, err := a.getKeys()
	if err != nil {
		return nil, fmt.Errorf("could not get keys")
	}

	pubKeyMap := make(map[string]interface{})
	equal := false
	for _, key := range keys {
		keyKidString := key.(map[string]interface{})["kid"].(string)
		if keyKidString == kid {
			equal = true
			pubKeyMap = key.(map[string]interface{})
		}
	}
	if !equal {
		return nil, fmt.Errorf("kid is not equal")
	}

	rawN := pubKeyMap["n"].(string)
	rawE := pubKeyMap["e"].(string)
	decodedE, err := base64.RawURLEncoding.DecodeString(rawE)
	if err != nil {
		return nil, err
	}

	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}
	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}
	decodedN, err := base64.RawURLEncoding.DecodeString(rawN)
	if err != nil {
		return nil, err
	}
	pubKey.N.SetBytes(decodedN)

	return pubKey, nil
}

func (a *Auth) getKeys() (keys []interface{}, err error) {
	response, err := http.Get(a.CognitoURL)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		err = errors.New(string(body))
		return nil, err
	}

	res := make(map[string]interface{})

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	keys = res["keys"].([]interface{})
	return
}
