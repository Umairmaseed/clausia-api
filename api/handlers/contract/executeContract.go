package contract

import (
	"github.com/google/logger"
	"github.com/umairmaseed/clausia-api/chaincode"
)

func ExecuteContract() {

	contracts, err := chaincode.GetExecutableContract()
	if err != nil {
		logger.Errorf("failed to get executable contracts: %v", err)
		return
	}

	for _, contractMap := range contracts {
		reqMap := map[string]interface{}{
			"contract": contractMap,
		}

		_, err := chaincode.ExecuteContract(reqMap)
		if err != nil {
			logger.Errorf("failed to execute contract: %v", err)
			continue
		}
	}
}
