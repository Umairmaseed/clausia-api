package contract

import (
	"crypto/sha256"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type addInputsToMakePaymentType struct {
	Clause              map[string]interface{} `form:"clause" binding:"required"`
	Date                string                 `form:"date" binding:"required"`
	StripeToken         string                 `form:"stripeToken"`
	PayPalTransactionID string                 `form:"payPalTransactionID"`
	Payment             float64                `form:"payment" binding:"required"`
	Receipt             *multipart.FileHeader  `form:"Receipt"`
	FinalPayment        bool                   `form:"finalPayment" binding:"required"`
}

func AddInputsToMakePayment(c *gin.Context) {
	var form addInputsToMakePaymentType
	if err := c.Bind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind request form: ", http.StatusBadRequest)
		return
	}

	reqMap := map[string]interface{}{
		"clause":       form.Clause,
		"payment":      form.Payment,
		"finalPayment": form.FinalPayment,
		"date":         form.Date,
	}

	if form.Receipt != nil {
		if form.PayPalTransactionID == "" && form.StripeToken == "" {

			fbytes, err := utils.GetFileBytes(form.Receipt)
			if err != nil {
				errorhandler.ReturnError(c, err, "Failed to read file: ", http.StatusInternalServerError)
				return
			}

			hash := fmt.Sprintf("%x", sha256.Sum256(fbytes))

			s3Url, err := utils.UploadReceiptToS3(fbytes, hash)
			if err != nil {
				logger.Error(err)
				c.String(http.StatusInternalServerError, "failed to upload file to s3: "+err.Error())
				c.Abort()
				return
			}

			reqMap["receiptUrl"] = s3Url
			reqMap["receiptHash"] = hash
		} else {
			errorhandler.ReturnError(c, fmt.Errorf("receipt cannot be provided with Stripe or PayPal payment"), "Invalid input", http.StatusBadRequest)
			return
		}
	}

	if form.StripeToken != "" {
		reqMap["stripeToken"] = form.StripeToken
	}
	if form.PayPalTransactionID != "" {
		reqMap["payPalTransactionID"] = form.PayPalTransactionID
	}

	updatedClause, err := chaincode.AddInputsToMakePayment(reqMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to add inputs to clause", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"clause": updatedClause})
}
