package chaincode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func GetExpectedUserDoc(reqMap map[string]interface{}) ([]map[string]interface{}, error) {
	path := os.Getenv("ORG_URL") + "/invoke/expectedUserDoc"

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
		return nil, fmt.Errorf("failed to change the status of the document to cancel")
	}

	var response struct {
		Result []map[string]interface{} `json:"result"`
	}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return response.Result, nil
}
