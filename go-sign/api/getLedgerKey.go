package api

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/go-sign/logger"
	"github.com/goledgerdev/go-sign/pdfsign"
)

type fileForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type responseObject struct {
	LedgerKey string `json:"ledgerKey"`
}

func GetLedgerKey(c *gin.Context) {
	var form fileForm
	log := logger.Logger().Sugar()
	if err := c.Bind(&form); err != nil {
		log.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	openedFile, err := form.File.Open()
	if err != nil {
		log.Error("Could not open file from form", err)
		c.String(400, "missing pdf file")
		return
	}

	pdfByte, err := io.ReadAll(openedFile)
	if err != nil {
		log.Error("failed to read pdf bytes: ", err)
		c.String(http.StatusInternalServerError, "failed to open PDF")
		return
	}

	ledgerKey := ""
	ledgerKey, err = pdfsign.RetrieveKey(pdfByte)
	if err != nil {
		log.Infof("could not retrieve ledgerKey form PDF: %s", err.Error())
	}

	log.Info("ledgerKey is: ", ledgerKey)

	var response responseObject
	response.LedgerKey = ledgerKey

	c.JSON(200, response)
}
