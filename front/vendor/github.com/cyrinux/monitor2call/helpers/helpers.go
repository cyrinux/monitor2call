package helpers

import "os"

// GetEnv get ENV var wrapper
// https://stackoverflow.com/questions/40326540/golang-how-to-assign-default-value-if-env-var-is-empty
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
