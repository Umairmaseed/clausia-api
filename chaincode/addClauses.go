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

func AddClauses(reqMap map[string]interface{}) (map[string]interface{}, error) {
	path := os.Getenv("ORG_URL") + "/invoke/addClauses"

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
		return nil, fmt.Errorf("failed to add clauses to the contract")
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
