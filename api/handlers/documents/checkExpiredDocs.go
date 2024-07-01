package documents

import (
	"encoding/json"

	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/google/logger"
)

func CheckExpiredDocs() {
	expiredDocs, err := chaincode.GetExpiredDocument()
	if err != nil {
		logger.Error(err)
	}

	for _, docMap := range expiredDocs {
		if status, ok := docMap["status"].(float64); ok && status == 0 {
			docMap["status"] = 2

			docJSON, err := json.Marshal(docMap)
			if err != nil {
				logger.Error("Failed to marshal document: ", err)
				continue
			}

			var fileAsset chaincode.FileAsset
			err = json.Unmarshal(docJSON, &fileAsset)
			if err != nil {
				logger.Error("Failed to unmarshal document: ", err)
				continue
			}
			_, err = chaincode.UploadDocumentTransaction(fileAsset)
			if err != nil {
				logger.Error(err)
			}
		}
	}

}
