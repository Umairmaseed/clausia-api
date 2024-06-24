package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cloudflare/cfssl/log"
)

func GetDoc(key string) (map[string]interface{}, error) {
	request, err := json.Marshal(map[string]interface{}{
		"key": map[string]interface{}{
			"@assetType": "document",
			"@key":       key,
		},
	})
	if err != nil {
		log.Error("Could not marshal request body", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	reqB64 := base64.StdEncoding.EncodeToString(request)

	getDoc := os.Getenv("ORG_URL") + "/query/getDoc?@request=" + reqB64

	res, err := http.Get(getDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("asset not found")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get file")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	result, ok := response["result"].([]interface{})
	if !ok || len(result) == 0 {
		return nil, errors.New("no result found in response")
	}

	return result[0].(map[string]interface{}), nil

}
