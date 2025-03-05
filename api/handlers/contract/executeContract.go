package contract

import (
	"github.com/goledgerdev/goprocess-api/chaincode"
	"github.com/google/logger"
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
