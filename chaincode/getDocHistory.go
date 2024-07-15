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

type DocumentHistoryRecord struct {
	TxID      string          `json:"txId"`
	Timestamp string          `json:"timestamp"`
	Value     json.RawMessage `json:"value"`
	IsDeleted bool            `json:"isDeleted"`
}

func GetDocHistory(key string) ([]DocumentHistoryRecord, error) {
	request, err := json.Marshal(map[string]interface{}{
		"key": map[string]interface{}{
			"@key": key,
		},
	})
	if err != nil {
		log.Error("Could not marshal request body", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	reqB64 := base64.StdEncoding.EncodeToString(request)

	getDoc := os.Getenv("ORG_URL") + "/query/getDocHistory?@request=" + reqB64

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
	var history []DocumentHistoryRecord
	err = json.Unmarshal(body, &history)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return history, nil
}
