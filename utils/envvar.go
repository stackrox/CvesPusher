package utils

import (
	"fmt"
	"os"
)

func IsEnvVarNonEmpty(env string) (string, error) {
	val, ok := os.LookupEnv(env)
	if !ok {
		return "", fmt.Errorf("environment variable %q is not set", env)
	}
	if val == "" {
		return "", fmt.Errorf("environment variable %q is empty", env)
	}
	return val, nil
}
