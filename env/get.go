package env

import (
	"fmt"
	"os"
)

func Get(varname string) (string, error) {
	value := os.Getenv(varname)
	if value == "" {
		return "", fmt.Errorf("missing %s env var", varname)
	}
	return value, nil
}
