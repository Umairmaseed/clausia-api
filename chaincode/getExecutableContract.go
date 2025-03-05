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

func GetExecutableContract() ([]map[string]interface{}, error) {
	path := os.Getenv("ORG_URL") + "/invoke/contractsWithExecutableClauses"

	// Creating an empty request map
	reqMap := map[string]interface{}{}
	body, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	requestBody := bytes.NewBuffer(body)

	res, err := http.Post(path, "application/json", requestBody)
	if err != nil {
		fmt.Println("error: " + err.Error())
		return nil, fmt.Errorf("failed to send request to chaincode: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get executable contracts, status code: %d", res.StatusCode)
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var resp []map[string]interface{}
	err = json.Unmarshal(responseBody, &resp)
	if err != nil {
		logger.Errorf("failed to unmarshal response from blockchain: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no contracts available with executable clauses to execute")
	}

	return resp, nil
}
