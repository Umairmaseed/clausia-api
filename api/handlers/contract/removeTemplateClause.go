package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/google/logger"
)

type RemoveTemplateClauseForm struct {
	Template       map[string]interface{} `form:"template" binding:"required"`
	TemplateClause map[string]interface{} `form:"templateClause" binding:"required"`
}

func RemoveTemplateClause(c *gin.Context) {
	var form RemoveTemplateClauseForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Error("Failed to bind request form: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := map[string]interface{}{
		"template":       form.Template,
		"templateClause": form.TemplateClause,
	}

	contract, err := chaincode.RemoveTemplateClause(req)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"templateClause": contract, "message": "template clause removed successfully"})

}
