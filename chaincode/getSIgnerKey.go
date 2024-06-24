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

func GetSignerKey(cpf string) (map[string]interface{}, error) {
	request, err := json.Marshal(map[string]interface{}{
		"cpf": cpf,
	})
	if err != nil {
		log.Error("Could not marshal request body", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	reqB64 := base64.StdEncoding.EncodeToString(request)

	getSignerURL := os.Getenv("ORG_URL") + "/query/getSignerKey?@request=" + reqB64

	res, err := http.Get(getSignerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("signer not found")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get signer")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return response, nil
}
