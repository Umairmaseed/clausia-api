package contract

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

type addInputsToCheckFine struct {
	Clause          map[string]interface{} `form:"clause" binding:"required"`
	ReferenceValue  *float64               `form:"referenceValue"`
	DailyPercentage *float64               `form:"dailyPercentage"`
	Days            *float64               `form:"days"`
}

func AddInputsToCheckFine(c *gin.Context) {
	var form addInputsToCheckFine
	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	if form.ReferenceValue == nil && form.DailyPercentage == nil && form.Days == nil {
		errorhandler.ReturnError(c, fmt.Errorf("no input values provided to update"), "No input values provided to update", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"clause": form.Clause,
	}

	if form.ReferenceValue != nil {
		reqMap["referenceValue"] = *form.ReferenceValue
	}
	if form.DailyPercentage != nil {
		reqMap["dailyPercentage"] = *form.DailyPercentage
	}
	if form.Days != nil {
		reqMap["days"] = *form.Days
	}

	updatedClause, err := chaincode.AddInputsToCheckFine(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add inputs to clause", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"clause": updatedClause})
}
