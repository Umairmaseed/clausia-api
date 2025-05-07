package documents

import (
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/chaincode"
)

func CheckExpiredDocs() {
	expiredDocs, err := chaincode.GetExpiredDocument()
	if err != nil {
		logger.Error(err)
	}

	for _, docMap := range expiredDocs {
		if status, ok := docMap["status"].(float64); ok && status == 0 {
			key := docMap["@key"].(string)
			docMap["status"] = 2

			logger.Infof(key)

			documentMAp := map[string]interface{}{
				"@assetType": "document",
				"@key":       key,
			}
			_, err = chaincode.UpdateDocument(documentMAp, docMap)
			if err != nil {
				logger.Error(err)
			}
		}
	}

}
