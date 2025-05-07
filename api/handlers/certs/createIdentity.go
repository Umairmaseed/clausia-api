package certs

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/umairmaseed/clausia-api/certs"
)

func CreateIdentityHandler(c *gin.Context, username string, commonName string, password string) ([]byte, error) {
	caMngr, err := certs.InitCAMngr(os.Getenv("SDK_CONFIG_PATH"), os.Getenv("CA_URL"))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to initialize CA manager"})
		return nil, err
	}

	pfx, err := caMngr.CreateIdentity(username, commonName, password)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create identity"})
		c.Abort()
		return nil, err
	}
	return pfx, nil
}
