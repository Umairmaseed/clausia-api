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

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
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
		errorhandler.ReturnError(c, err, "Failed to bind form data", http.StatusBadRequest)
		return
	}

	// Retrieving document from blockchain
	asset, err := chaincode.GetDoc(form.DocKey)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to retrieve document asset", http.StatusInternalServerError)
		return
	}

	originalDocURL, _ := asset["originalDocURL"].(string)
	fileName, _ := asset["name"].(string)
	status := asset["status"].(float64)
	originalHash := asset["originalHash"].(string)
	requiredSignatures, _ := asset["requiredSignatures"].([]interface{})
	successfulSignatures, _ := asset["successfulSignatures"].([]interface{})
	rejectedSignatures := asset["rejectedSignatures"].([]interface{})
	ownerMap, ok := asset["owner"].(map[string]interface{})
	if !ok {
		errorhandler.ReturnError(c, err, "Owner is not a valid map", http.StatusInternalServerError)
		return
	}
	ownerKey, ok := ownerMap["@key"].(string)
	if !ok {
		errorhandler.ReturnError(c, err, "Owner key not found", http.StatusInternalServerError)
		return
	}
	owner := chaincode.Signer{Key: ownerKey}
	username := form.Username

	finalDocURL, ok := asset["finalDocURL"].(string)
	if ok {
		retrieveOriginalDocURL = false
	}

	// Checking status before signing the doc
	if status == 1 {
		errorhandler.ReturnError(c, nil, "Document is not available for signatures", http.StatusInternalServerError)
		return
	} else if status == 2 {
		errorhandler.ReturnError(c, nil, "Document is expired to be signed", http.StatusInternalServerError)
		return
	} else if status == 3 || status == 4 {
		errorhandler.ReturnError(c, nil, "Document is already finalized for signatures", http.StatusInternalServerError)
		return
	}

	// Retrieving signer key and signer asset from blockchain
	signerKey, err := chaincode.GetSignerKey(form.Cpf)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to retrieve signer key", http.StatusInternalServerError)
		return
	}
	ledgerKey := signerKey["@key"].(string)

	_, err = chaincode.GetSigner(ledgerKey)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to retrieve signer asset", http.StatusInternalServerError)
		return
	}

	// Checking Signer eligible to sign the document
	signerAllowed := false
	for _, reqSigner := range requiredSignatures {
		reqSignerMap, _ := reqSigner.(map[string]interface{})
		if reqSignerMap["@key"] == ledgerKey {
			signerAllowed = true
			break
		}
	}

	if !signerAllowed {
		errorhandler.ReturnError(c, err, "Signer not allowed to sign the document", http.StatusForbidden)
		return
	}

	//Check if signer already signed the document
	for _, sig := range successfulSignatures {
		signerMap, _ := sig.(map[string]interface{})
		key, _ := signerMap["@key"].(string)
		if key == ledgerKey {
			errorhandler.ReturnError(c, err, "Document already signed by the signer", http.StatusForbidden)
			return
		}
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")

	// Validate which docurl should be used to retrieve the doc from s3
	var s3FilePath string
	if retrieveOriginalDocURL {
		s3FilePath = strings.TrimPrefix(originalDocURL, fmt.Sprintf("s3://%s/", bucketName))
	} else {
		s3FilePath = strings.TrimPrefix(finalDocURL, fmt.Sprintf("s3://%s/", bucketName))
	}

	//  Retrieve document from s3
	docBytes, err := utils.DownloadFileFromS3(c.Request.Context(), s3FilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to download document: "+err.Error())
		return
	}

	//Retrive certificate from from s3
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
		errorhandler.ReturnError(c, err, "Failed to create form file for document", http.StatusInternalServerError)
		return
	}
	fileWriter.Write(docBytes)

	certWriter, err := bodyWriter.CreateFormFile("certificate", fmt.Sprintf("%s_cert.pfx", username))
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to create form file for certificate", http.StatusInternalServerError)
		return
	}
	certWriter.Write(certBytes)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	// Signing the document and reading the response
	response, err := client.Post(url, contentType, bodyBuf)
	if err != nil {
		errorhandler.ReturnError(c, err, "Could not sign document due to error", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		errorhandler.ReturnError(c, err, "Unable to read response from signing service", http.StatusInternalServerError)
		return
	}

	var res signResponse
	if err := json.Unmarshal(resBody, &res); err != nil {
		errorhandler.ReturnError(c, err, "Could not parse response from signing service:", http.StatusInternalServerError)
		return
	}

	if res.Error != "" {
		errorhandler.ReturnError(c, err, res.Error, http.StatusInternalServerError)
		return
	}

	// Organizing the finalDocurl and saving it in the s3
	var finalHashName string
	signedDocHash := fmt.Sprintf("%x", sha256.Sum256(res.File))

	parts := strings.Split(fileName, "-")
	if len(parts) > 1 {
		finalHashName = signedDocHash + "-" + parts[1]
	}

	signedDocUrl, err := utils.UploadSignedDocToS3(res.File, finalHashName)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to upload document to S3:", http.StatusInternalServerError)
		return
	}

	//Organizing updateSignature, rejectedSignatures and requiredSignatures
	updatedSuccessfulSignatures := convertToSigners(successfulSignatures)

	// Append ledgerKey to updatedSuccessfulSignatures
	updatedSuccessfulSignatures = append(updatedSuccessfulSignatures, chaincode.Signer{Key: ledgerKey})

	requiredSigners := convertToSigners(requiredSignatures)

	var successfulSigners []chaincode.Signer
	for _, sig := range updatedSuccessfulSignatures {
		key := sig.Key
		successfulSigners = append(successfulSigners, chaincode.Signer{Key: key})
	}

	rejectedSigners := convertToSigners(rejectedSignatures)

	// Updating doc asset state
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
		Owner:                owner,
	})
	if err != nil {
		errorhandler.ReturnError(c, err, "failed to save document to ledger:", http.StatusInternalServerError)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, res)
}

func convertToSigners(signatures []interface{}) []chaincode.Signer {
	var signers []chaincode.Signer
	for _, sig := range signatures {
		signerMap, _ := sig.(map[string]interface{})
		key, _ := signerMap["@key"].(string)
		signers = append(signers, chaincode.Signer{Key: key})
	}
	return signers
}
