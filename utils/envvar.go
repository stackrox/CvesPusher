package utils

import (
	"fmt"
	"os"
)

func IsEnvVarNonEmpty(env string) (string, error) {
	val := os.Getenv(env)
	if val == "" {
		return "", fmt.Errorf("environment variable %q is either unset or empty", env)
	}
	return val, nil
}
