package documents

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
	"github.com/google/logger"
)

type signForm struct {
	DocKey    string `form:"dockey" binding:"required"`
	Password  string `form:"password" binding:"required"`
	Signature string `form:"signature" binding:"required"`
	Username  string `form"username" binding:"required"`
	Cpf       string `form"cpf" binding:"required"`
}

type signResponse struct {
	Name          string `json:"name"`
	File          []byte `json:"file"`
	PubKey        string `json:"pubKey"`
	QRCode        string `json:"qrCode"`
	SaltedHash    string `json:"saltedHash"`
	LastSignature string `json:"lastSignature"`
	Key           string `json:"key"`
	Error         string `json:"error"`
}

func SignDocument(c *gin.Context) {
	var form signForm
	var retrieveOriginalDocURL bool = true
	if err := c.ShouldBind(&form); err != nil {
		log.Error("Failed to bind form data", err)
		c.String(http.StatusBadRequest, "Failed to bind form data: "+err.Error())
		return
	}

	asset, err := chaincode.GetDoc(form.DocKey)
	if err != nil {
		log.Error("Failed to retrieve document asset", err)
		c.String(http.StatusInternalServerError, "Failed to retrieve document asset: "+err.Error())
		return
	}

	originalDocURL, ok := asset["originalDocURL"].(string)
	if !ok {
		log.Error("Invalid document asset: missing filename")
		c.String(http.StatusInternalServerError, "Invalid document asset: missing filename")
		return
	}

	fileName, ok := asset["name"].(string)
	if !ok {
		log.Error("Failed to get the name of the asset")
		c.String(http.StatusInternalServerError, "Failed to get the name of the asset")
		return
	}

	status := asset["status"].(float64)
	originalHash := asset["originalHash"].(string)

	if status != 0 {
		log.Error("Document with status waiting can be signed only")
		c.String(http.StatusInternalServerError, "Document with status waiting can be signed only")
		return
	}

	finalDocURL, ok := asset["finalDocURL"].(string)
	if ok {
		retrieveOriginalDocURL = false
	}

	username := form.Username

	signerKey, err := chaincode.GetSignerKey(form.Cpf)
	if err != nil {
		log.Error("Failed to retrieve signer key", err)
		c.String(http.StatusInternalServerError, "Failed to retrieve signer key: "+err.Error())
		return
	}

	requiredSignatures, _ := asset["requiredSignatures"].([]interface{})
	successfulSignatures, _ := asset["successfulSignatures"].([]interface{})
	rejectedSignatures := asset["rejectedSignatures"].([]interface{})
	ledgerKey := signerKey["@key"].(string)

	_, err = chaincode.GetSigner(ledgerKey)
	if err != nil {
		log.Error("Failed to retrieve signer asset", err)
		c.String(http.StatusInternalServerError, "Failed to retrieve signer asset: "+err.Error())
		return
	}

	signerAllowed := false
	for _, reqSigner := range requiredSignatures {
		reqSignerMap, ok := reqSigner.(map[string]interface{})
		if !ok {
			log.Error("Invalid required signer type")
			c.String(http.StatusInternalServerError, "Invalid required signer type")
			return
		}
		if reqSignerMap["@key"] == ledgerKey {
			signerAllowed = true
			break
		}
	}

	if !signerAllowed {
		log.Error("Signer not allowed to sign the document")
		c.String(http.StatusForbidden, "Signer not allowed to sign the document")
		return
	}

	for _, sig := range successfulSignatures {
		signerMap, ok := sig.(map[string]interface{})
		if !ok {
			log.Error("Invalid signer type")
			c.String(http.StatusInternalServerError, "Invalid signer type")
			return
		}
		key, ok := signerMap["@key"].(string)
		if !ok {
			log.Error("Invalid signer key type")
			c.String(http.StatusInternalServerError, "Invalid signer key type")
			return
		}
		if key == ledgerKey {
			log.Error("Document already signed by the signer")
			c.String(http.StatusForbidden, "Document already signed by the signer")
			return
		}
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")

	var s3FilePath string
	if retrieveOriginalDocURL {
		s3FilePath = strings.TrimPrefix(originalDocURL, fmt.Sprintf("s3://%s/", bucketName))
	} else {
		s3FilePath = strings.TrimPrefix(finalDocURL, fmt.Sprintf("s3://%s/", bucketName))
	}

	docBytes, err := utils.DownloadFileFromS3(c.Request.Context(), s3FilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to download document: "+err.Error())
		return
	}
	certKey := fmt.Sprintf("certificates/%s_cert.pfx", username)

	certBytes, err := utils.DownloadFileFromS3(c.Request.Context(), certKey)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to download certificate: "+err.Error())
		return
	}
	url := fmt.Sprintf("%s/api/signdocs", os.Getenv("GO_SIGN_API"))
	client := http.DefaultClient

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("fileName", fileName)
	bodyWriter.WriteField("password", form.Password)
	bodyWriter.WriteField("signature", form.Signature)
	bodyWriter.WriteField("clientBaseUrl", os.Getenv("CLIENT_BASE_URL"))
	bodyWriter.WriteField("ledgerKey", ledgerKey)

	fileWriter, err := bodyWriter.CreateFormFile("file", fileName)
	if err != nil {
		log.Error("Failed to create form file for document", err)
		c.String(http.StatusInternalServerError, "Failed to create form file for document: "+err.Error())
		return
	}
	fileWriter.Write(docBytes)

	certWriter, err := bodyWriter.CreateFormFile("certificate", fmt.Sprintf("%s_cert.pfx", username))
	if err != nil {
		log.Error("Failed to create form file for certificate", err)
		c.String(http.StatusInternalServerError, "Failed to create form file for certificate: "+err.Error())
		return
	}
	certWriter.Write(certBytes)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	response, err := client.Post(url, contentType, bodyBuf)
	if err != nil {
		log.Error("Could not sign document due to error", err)
		c.String(http.StatusInternalServerError, "Could not connect to the signing service: "+err.Error())
		return
	}
	defer response.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("Unable to read response", err)
		c.String(http.StatusInternalServerError, "Unable to read response from signing service: "+err.Error())
		return
	}

	var res signResponse
	if err := json.Unmarshal(resBody, &res); err != nil {
		log.Error("Could not unmarshal response body", err)
		c.String(http.StatusInternalServerError, "Could not parse response from signing service: "+err.Error())
		return
	}

	if res.Error != "" {
		log.Error(res.Error)
		c.String(http.StatusInternalServerError, "Failed to sign document: "+res.Error)
		return
	}
	var finalHashName string
	signedDocHash := fmt.Sprintf("%x", sha256.Sum256(res.File))

	parts := strings.Split(fileName, "-")
	if len(parts) > 1 {
		finalHashName = signedDocHash + "-" + parts[1]
	}

	signedDocUrl, err := utils.UploadSignedDocToS3(res.File, finalHashName)
	if err != nil {
		log.Error("Failed to upload file to S3", err)
		c.String(http.StatusInternalServerError, "Failed to upload document to S3: "+err.Error())
		return
	}

	var updatedSuccessfulSignatures []interface{}
	for _, sig := range successfulSignatures {
		signerMap, ok := sig.(map[string]interface{})
		if !ok {
			log.Error("Invalid signer type")
			c.String(http.StatusInternalServerError, "Invalid signer type")
			return
		}
		key, ok := signerMap["@key"].(string)
		if !ok {
			log.Error("Invalid signer key type")
			c.String(http.StatusInternalServerError, "Invalid signer key type")
			return
		}
		updatedSuccessfulSignatures = append(updatedSuccessfulSignatures, key)
	}

	// Append ledgerKey to updatedSuccessfulSignatures
	updatedSuccessfulSignatures = append(updatedSuccessfulSignatures, ledgerKey)

	var requiredSigners []chaincode.Signer
	for _, sig := range requiredSignatures {
		signerMap, ok := sig.(map[string]interface{})
		if !ok {
			log.Error("Invalid signer type")
			c.String(http.StatusInternalServerError, "Invalid signer type")
			return
		}
		key, ok := signerMap["@key"].(string)
		if !ok {
			log.Error("Invalid signer key type")
			c.String(http.StatusInternalServerError, "Invalid signer key type")
			return
		}
		requiredSigners = append(requiredSigners, chaincode.Signer{Key: key})
	}

	var successfulSigners []chaincode.Signer
	for _, sig := range updatedSuccessfulSignatures {
		key, ok := sig.(string)
		if !ok {
			log.Error("Invalid signer key type")
			c.String(http.StatusInternalServerError, "Invalid signer key type")
			return
		}
		successfulSigners = append(successfulSigners, chaincode.Signer{Key: key})
	}

	var rejectedSigners []chaincode.Signer
	for _, sig := range rejectedSignatures {
		signerMap, ok := sig.(map[string]interface{})
		if !ok {
			log.Error("Invalid signer type")
			c.String(http.StatusInternalServerError, "Invalid signer type")
			return
		}
		key, ok := signerMap["@key"].(string)
		if !ok {
			log.Error("Invalid signer key type")
			c.String(http.StatusInternalServerError, "Invalid signer key type")
			return
		}
		rejectedSigners = append(rejectedSigners, chaincode.Signer{Key: key})
	}

	_, err = chaincode.UploadDocumentTransaction(chaincode.FileAsset{
		OriginalHash:         originalHash,
		Status:               int(status),
		RequiredSignatures:   requiredSigners,
		OriginalDocURL:       originalDocURL,
		Name:                 fileName,
		RejectedSignatures:   rejectedSigners,
		SuccessfulSignatures: successfulSigners,
		FinalHash:            signedDocHash,
		FinalDocURL:          signedDocUrl,
	})
	if err != nil {
		logger.Error(err)
		c.String(http.StatusInternalServerError, "failed to save document to ledger: "+err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, res)
}
