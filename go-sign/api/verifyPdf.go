package api

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/go-sign/api/middleware"
	"github.com/goledgerdev/go-sign/logger"
	"github.com/goledgerdev/go-sign/pdfsign"
)

type verifyForm struct {
	Files []*multipart.FileHeader `form:"files" binding:"required"`

	// this should be an map with filename/password
	Passwords string `form:"passwords"`
}

type PasswordMap map[string]string

type verifyResponse struct {
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
	Error string `json:"error"`
	Key   string `json:"key"`
}

// VerifyPDF returns the total amount of certificates
func VerifyPDF(c *gin.Context) {
	var form verifyForm
	log := logger.Logger().Sugar()
	if err := c.Bind(&form); err != nil {
		log.Error(err)
		return
	}

	pwd := make(PasswordMap)
	if form.Passwords != "" {
		err := json.Unmarshal([]byte(form.Passwords), &pwd)
		if err != nil {
			log.Error(err)
			c.Error(middleware.NewErrStr(http.StatusBadRequest, "Could not decode signature info from body"))
			return
		}
	}

	pdfs := make([]pdfsign.PDFInput, len(form.Files))
	errorsResult := make([]string, len(form.Files))
	for i, formFile := range form.Files {
		openedFile, err := formFile.Open()
		if err != nil {
			errorsResult[i] = err.Error()
			continue
		}

		pdfs[i] = pdfsign.PDFInput{
			Filename: formFile.Filename,
			Pdf:      openedFile,
		}
		password, ok := pwd[formFile.Filename]
		if ok {
			pdfs[i].Password = password
		}
	}

	var response = struct {
		Files []verifyResponse `json:"files"`
	}{}
	verifyResults := pdfsign.VerifyDocument(pdfs...)

	for _, res := range verifyResults {
		response.Files = append(response.Files, verifyResponse{
			Name:  strings.TrimSuffix(res.Filename, filepath.Ext(res.Filename)),
			Valid: res.Valid,
			Error: res.Error,
			Key:   res.Key,
		})
	}
	c.JSON(200, response)
}
