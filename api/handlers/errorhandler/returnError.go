package errorhandler

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
)

func ReturnError(c *gin.Context, err error, errorText string, status int) {
	log.Error(errorText)
	if err != nil {
		c.String(status, "%s: %s", errorText, err.Error())
	} else {
		c.String(status, errorText)
	}
}
