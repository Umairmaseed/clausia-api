package api

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/umairmaseed/go-sign/logger"
)

type saltForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type saltedFileResponse struct {
	File       []byte `json:"file"`
	SaltedHash string `json:"saltedHash"`
}

func SaltPdf(c *gin.Context) {
	var form saltForm
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
		c.String(http.StatusInternalServerError, "falha na abertura do PDF")
		return
	}

	u := uuid.New().String()
	saltedPdf := append(pdfByte, []byte(u)...)

	pdfHash := sha256.Sum256([]byte(saltedPdf))
	saltedPdfDigest := fmt.Sprintf("%x", pdfHash[:])

	response := saltedFileResponse{
		File:       saltedPdf,
		SaltedHash: saltedPdfDigest,
	}

	c.JSON(200, response)
}
