package documents

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goledgerdev/goprocess-api/api/handlers/errorhandler"
	"github.com/goledgerdev/goprocess-api/chaincode"
)

type DocumentHistoryEntry struct {
	TxID      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	Signer    string    `json:"signer"`
}

func GetDoc(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		errorhandler.ReturnError(c, fmt.Errorf("missing key"), "Key is required", http.StatusBadRequest)
		return
	}

	asset, err := chaincode.GetDoc(key)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to retrieve document asset", http.StatusInternalServerError)
		return
	}

	history, err := chaincode.GetDocHistory(key)
	if err != nil {
		errorhandler.ReturnError(c, err, "Failed to retrieve document history", http.StatusInternalServerError)
		return
	}

	filteredHistory := filterSuccessfulSigners(history)

	c.JSON(http.StatusOK, gin.H{
		"document":        asset,
		"filteredHistory": filteredHistory,
	})

}

func filterSuccessfulSigners(history []chaincode.DocumentHistoryRecord) []DocumentHistoryEntry {
	var filteredHistory []DocumentHistoryEntry
	previousSigners := make(map[string]bool)

	for _, entry := range history {
		var doc chaincode.FileAsset
		err := json.Unmarshal(entry.Value, &doc)
		if err != nil {
			continue
		}

		secondsStr := extractValue(entry.Timestamp, "seconds:")
		nanosStr := extractValue(entry.Timestamp, "nanos:")

		// Convert extracted values to integers
		seconds, err := strconv.ParseInt(secondsStr, 10, 64)
		if err != nil {
			fmt.Println("Error parsing seconds:", err)
		}

		nanos, err := strconv.ParseInt(nanosStr, 10, 64)
		if err != nil {
			fmt.Println("Error parsing nanos:", err)
		}

		// Create a time object using Unix time
		t := time.Unix(seconds, nanos)

		for _, signer := range doc.SuccessfulSignatures {
			signerKey := signer.Key
			if !previousSigners[signerKey] {
				filteredHistory = append(filteredHistory, DocumentHistoryEntry{
					TxID:      entry.TxID,
					Timestamp: t,
					Signer:    signerKey,
				})
				previousSigners[signerKey] = true
			}
		}
	}

	return filteredHistory
}

func extractValue(input, prefix string) string {
	startIndex := strings.Index(input, prefix)
	if startIndex == -1 {
		return ""
	}
	startIndex += len(prefix)
	endIndex := strings.Index(input[startIndex:], " ")
	if endIndex == -1 {
		endIndex = len(input)
	} else {
		endIndex += startIndex
	}
	return input[startIndex:endIndex]
}
