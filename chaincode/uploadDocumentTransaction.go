package chaincode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/logger"
)

func UploadDocumentTransaction(f FileAsset) (map[string]interface{}, error) {
	f.AssetType = "document"

	successfulSignatureSlice := []Signer{}
	successfulSignatureSlice = append(successfulSignatureSlice, f.SuccessfulSignatures...)

	rejectedSignatureSlice := []Signer{}
	rejectedSignatureSlice = append(rejectedSignatureSlice, f.RejectedSignatures...)

	path := os.Getenv("ORG_URL") + "/invoke/uploadDocument"
	reqMap := map[string]interface{}{
		"originalHash":         f.OriginalHash,
		"status":               f.Status,
		"requiredSignatures":   f.RequiredSignatures,
		"originalDocURL":       f.OriginalDocURL,
		"name":                 f.Name,
		"rejectedSignatures":   rejectedSignatureSlice,
		"successfulSignatures": successfulSignatureSlice,
		"owner":                f.Owner,
		"timeout":              f.Timeout,
	}
	if f.FinalHash != "" {
		reqMap["finalHash"] = f.FinalHash
	}
	if f.FinalDocURL != "" {
		reqMap["finalDocURL"] = f.FinalDocURL
	}
	if f.Signature.Key != "" {
		reqMap["signature"] = f.Signature
	}

	body, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	requestBody := bytes.NewBuffer(body)

	res, err := http.Post(path, "application/json", requestBody)
	if err != nil {
		fmt.Println("error: " + err.Error())
		fmt.Println("res: ", res)
		return nil, fmt.Errorf("failed to send request to chaincode: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		fmt.Println("res: ", res)
		return nil, fmt.Errorf("failed to create asset")
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var resp map[string]interface{}
	err = json.Unmarshal(responseBody, &resp)
	if err != nil {
		logger.Errorf("failed to unmarshal response from blockchain")
	}

	return resp, nil
}
