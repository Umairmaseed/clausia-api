package documents

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/goledgerdev/goprocess-api/utils"
)

type docUrl struct {
	OriginalUrl *string `json:"originalurl" form:"originalurl" `
	FinalUrl    *string `json:"finalurl" form:"finalurl" `
}

func DownloadDocument(c *gin.Context) {
	var form docUrl

	if err := c.ShouldBind(&form); err != nil {
		errorhandler.ReturnError(c, err, "Failed to bind form data", http.StatusBadRequest)
		return
	}

	if (form.OriginalUrl == nil && form.FinalUrl == nil) || (form.OriginalUrl != nil && form.FinalUrl != nil) {
		errorhandler.ReturnError(c, fmt.Errorf("either originalurl or finalurl must be provided, but not both"), "Invalid input: provide either originalurl or finalurl", http.StatusBadRequest)
		return
	}

	queryMap := map[string]interface{}{
		"@assetType": "document",
	}

	if form.OriginalUrl != nil {
		queryMap["originalDocURL"] = *form.OriginalUrl
	}

	if form.FinalUrl != nil {
		queryMap["finalDocURL"] = *form.FinalUrl
	}

	docAsset, err := chaincode.SearchAsset(queryMap)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to search for documents", http.StatusInternalServerError)
		return
	}
	resultArray, ok := docAsset["result"].([]interface{})
	if !ok || len(resultArray) == 0 {
		errorhandler.ReturnError(c, fmt.Errorf("no document found with the provided URL"), "No document found with the provided URL", http.StatusNotFound)
		return
	}

	document := resultArray[0].(map[string]interface{})
	name := document["name"].(string)
	url := ""
	if form.OriginalUrl != nil {
		url = *form.OriginalUrl
	} else {
		url = *form.FinalUrl
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	s3FilePath := strings.TrimPrefix(url, fmt.Sprintf("s3://%s/", bucketName))

	docBytes, err := utils.DownloadFileFromS3(c.Request.Context(), s3FilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to download document: "+err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+name)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", len(docBytes)))

	// Send the file
	c.Data(http.StatusOK, "application/octet-stream", docBytes)

}
