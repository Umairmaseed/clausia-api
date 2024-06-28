package chaincode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func GetExpiredDocument() ([]map[string]interface{}, error) {

	getDoc := os.Getenv("ORG_URL") + "/query/getExpiredDocuments"

	res, err := http.Get(getDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get expired documents")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response []map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if len(response) == 0 {
		return nil, nil
	}

	return response, nil

}
