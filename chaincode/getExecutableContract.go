package chaincode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/logger"
)

func GetExecutableContract() (map[string]interface{}, error) {
	path := os.Getenv("ORG_URL") + "/invoke/contractsWithExecutableClauses"

	res, err := http.Get(path)
	if err != nil {
		fmt.Println("error: " + err.Error())
		fmt.Println("res: ", res)
		return nil, fmt.Errorf("failed to send request to chaincode: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		fmt.Println("res: ", res)
		return nil, fmt.Errorf("failed to get executable contracts")
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
