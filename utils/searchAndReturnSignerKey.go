package utils

import (
	"fmt"

	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/chaincode"
)

func SearchAndReturnSignerKey(email string) (string, error) {
	queryMap := map[string]interface{}{
		"@assetType": "user",
		"email":      email,
	}

	signerAsset, err := chaincode.SearchAsset(queryMap)
	if err != nil {
		logger.Error(err)
		return "", fmt.Errorf("failed to search for user: %w", err)
	}

	resultArray, ok := signerAsset["result"].([]interface{})
	if !ok || len(resultArray) == 0 {
		logger.Error("Signer asset result is empty or not in expected format")
		return "", fmt.Errorf("signer asset result is empty or not in expected format")
	}

	firstResult := resultArray[0].(map[string]interface{})
	signerKey, ok := firstResult["@key"].(string)
	if !ok {
		logger.Error("Unable to extract @key from signer asset result")
		return "", fmt.Errorf("unable to extract @key from signer asset result")
	}
	return signerKey, nil
}
