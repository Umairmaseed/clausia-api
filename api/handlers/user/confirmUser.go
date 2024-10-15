package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/auth"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type GetUserForm struct {
	UserName string `form:"userName" `
	Email    string `form:"email" `
	Id       string `form:"id" `
}

func ConfirmUser(c *gin.Context) {
	var form GetUserForm

	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var userId string
	var err error

	if form.Email != "" {
		userId, err = findUserByEmail(form.Email)
	} else if form.UserName != "" {
		userId, err = findUserByUserName(form.UserName)
	} else if form.Id != "" {
		userId, err = findUserByID(form.Id)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No parameter was passed"})
		return
	}

	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User found",
		"userId":  userId,
	})
}

func findUserByEmail(email string) (string, error) {
	key, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		return "", err
	}
	return key, nil
}

func findUserByUserName(userName string) (string, error) {
	auth := auth.NewAuth()
	_, email, err := auth.CheckIfUserExistsAndGetEmail(userName)
	if err != nil {
		return "", err
	}

	key, nerr := utils.SearchAndReturnSignerKey(email)
	if nerr != nil {
		return "", err
	}
	return key, nil
}

func findUserByID(id string) (string, error) {
	_, err := chaincode.GetSigner(id)
	if err != nil {
		return "", err
	}
	return id, nil
}
