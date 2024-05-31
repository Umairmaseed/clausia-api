package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Auth) SignOut(c *gin.Context) {
	c.SetCookie("idToken", "", -1, "/", "", false, false)
	c.SetCookie("accessToken", "", -1, "/", "", false, false)
	c.SetCookie("refreshToken", "", -1, "/", "", false, false)

	c.Status(http.StatusOK)
}
