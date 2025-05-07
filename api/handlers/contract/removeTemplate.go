package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/chaincode"
)

type RemoveTemplateForm struct {
	Template map[string]interface{} `form:"template" binding:"required"`
}

func RemoveTemplate(c *gin.Context) {
	var form RemoveTemplateForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Error("Failed to bind request form: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := map[string]interface{}{
		"template": form.Template,
	}

	contract, err := chaincode.RemoveTemplate(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": contract, "message": "template removed successfully"})

}
