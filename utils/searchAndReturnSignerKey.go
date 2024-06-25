package utils

import (
	"fmt"

	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/google/logger"
)

func SearchAndReturnSignerKey(email string) (string, error) {
	queryMap := map[string]interface{}{
		"@assetType": "signer",
		"email":      email,
	}

	signerAsset, err := chaincode.SearchAsset(queryMap)
	if err != nil {
		logger.Error(err)
		return "", fmt.Errorf("failed to search for signer: %w", err)
	}

	resultArray, ok := signerAsset["result"].([]interface{})
	if !ok || len(resultArray) == 0 {
		logger.Error("Signer asset result is empty or not in expected format")
		return "", fmt.Errorf("Signer asset result is empty or not in expected format")
	}

	firstResult := resultArray[0].(map[string]interface{})
	signerKey, ok := firstResult["@key"].(string)
	if !ok {
		logger.Error("Unable to extract @key from signer asset result")
		return "", fmt.Errorf("Unable to extract @key from signer asset result")
	}
	return signerKey, nil
}
