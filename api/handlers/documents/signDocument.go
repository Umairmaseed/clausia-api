package documents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/cloudflare/cfssl/log"
	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/utils"
)

type signForm struct {
	Filename  string `form:"filename" binding:"required"`
	Password  string `form:"password" binding:"required"`
	Signature string `form:"signature" binding:"required"`
	Username  string `form"username" binding:"required"`
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
	if err := c.ShouldBind(&form); err != nil {
		log.Error("Failed to bind form data", err)
		c.String(http.StatusBadRequest, "Failed to bind form data: "+err.Error())
		return
	}

	username := form.Username

	s3FilePath := "/documents/" + form.Filename

	docBytes, err := utils.DownloadFileFromS3(c.Request.Context(), s3FilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to download document: "+err.Error())
		return
	}
	certKey := fmt.Sprintf("/certificates/%s_cert.pfx", username)

	certBytes, err := utils.DownloadFileFromS3(c.Request.Context(), certKey)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to download certificate: "+err.Error())
		return
	}
	url := fmt.Sprintf("%s/api/signdocs", os.Getenv("GO_SIGN_API"))
	client := http.DefaultClient

	// this should be changed later
	ledgerKey := "document:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("fileName", form.Filename)
	bodyWriter.WriteField("password", form.Password)
	bodyWriter.WriteField("signature", form.Signature)
	bodyWriter.WriteField("clientBaseUrl", os.Getenv("CLIENT_BASE_URL"))
	bodyWriter.WriteField("ledgerKey", ledgerKey)

	fileWriter, err := bodyWriter.CreateFormFile("file", form.Filename)
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

	_, err = utils.UploadSignedDocToS3(res.File, form.Filename)
	if err != nil {
		log.Error("Failed to upload file to S3", err)
		c.String(http.StatusInternalServerError, "Failed to upload document to S3: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}
