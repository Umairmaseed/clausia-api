package documents

import (
	"crypto/sha256"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/api/handlers/errorhandler"
	"github.com/umairmaseed/clausia-api/chaincode"
	"github.com/umairmaseed/clausia-api/db"
	"github.com/umairmaseed/clausia-api/utils"
)

type uploadDocumentForm struct {
	Files              []*multipart.FileHeader `form:"files" binding:"required"`
	RequiredSignatures string                  `form:"requiredSignatures" binding:"required"`
	Timeout            string                  `form:"timeout" binding:"required"`
}

func UploadDocument(c *gin.Context) {
	var form uploadDocumentForm

	if err := c.Bind(&form); err != nil {
		logger.Error(err)
		c.String(http.StatusBadRequest, err.Error())
	}

	timeout := form.Timeout

	email := c.Request.Header.Get("Email")
	if email == "" {
		logger.Error("Email not found in headers")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in headers"})
		return
	}

	signerKey, err := utils.SearchAndReturnSignerKey(email)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ownerMap := chaincode.Signer{
		Key: signerKey,
	}

	fileHashes := make([]string, len(form.Files))

	var response interface{}
	for _, f := range form.Files {

		fbytes, err := utils.GetFileBytes(f)
		if err != nil {
			logger.Error(err)
			c.String(http.StatusInternalServerError, "failed to get file bytes: "+err.Error())
			c.Abort()
			return
		}
		requiredSignatures, _ := parseRequiredSignatures(form.RequiredSignatures)
		if err := validateRequiredSignatures(requiredSignatures); err != nil {
			logger.Error(err)
			c.String(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		hash := fmt.Sprintf("%x", sha256.Sum256(fbytes))
		filename := f.Filename
		s3Url, err := utils.UploadFileToS3(fbytes, filename)
		if err != nil {
			logger.Error(err)
			c.String(http.StatusInternalServerError, "failed to upload file to s3: "+err.Error())
			c.Abort()
			return
		}

		processAsset, err := chaincode.UploadDocumentTransaction(chaincode.FileAsset{
			OriginalHash:       hash,
			Status:             0,
			RequiredSignatures: requiredSignatures,
			OriginalDocURL:     s3Url,
			Name:               filename,
			Owner:              ownerMap,
			Timeout:            timeout,
		})
		if err != nil {
			logger.Error(err)
			c.String(http.StatusInternalServerError, "failed to save document to ledger: "+err.Error())
			c.Abort()
			return
		}

		fileHashes = append(fileHashes, hash)

		response = processAsset
	}

	notification := []db.Notification{
		{
			UserID:  signerKey,
			Type:    "document",
			Message: "Document uploaded successfully",
		},
		{
			UserID:  form.RequiredSignatures,
			Type:    "document",
			Message: "You have been request by" + signerKey + " to sign a document",
		},
	}

	_, err = db.NewNotificationService(db.GetDB().Database()).CreateNotification(c.Request.Context(), &notification)
	if err != nil {
		errorhandler.ReturnError(c, err, "failed to generate notification", http.StatusInternalServerError)
	}

	c.Set("fileHashes", fileHashes)

	c.JSON(http.StatusCreated, response)

}

func parseRequiredSignatures(signatures string) ([]chaincode.Signer, error) {
	signerStrings := strings.Split(signatures, ",")
	requiredSignatures := make([]chaincode.Signer, 0, len(signerStrings))
	for _, signerStr := range signerStrings {
		signerStr = strings.TrimSpace(signerStr)
		if signerStr == "" {
			return nil, fmt.Errorf("invalid requiredSignatures format: empty signer detected")
		}
		requiredSignatures = append(requiredSignatures, chaincode.Signer{Key: signerStr})
	}
	return requiredSignatures, nil
}

func validateRequiredSignatures(signatures []chaincode.Signer) error {
	if len(signatures) == 0 {
		return fmt.Errorf("requiredSignatures cannot be empty")
	}
	for _, signer := range signatures {
		if signer.Key == "" {
			return fmt.Errorf("invalid requiredSignatures format: signer key cannot be empty")
		}
	}
	return nil
}
