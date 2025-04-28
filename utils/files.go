package utils

import (
	"encoding/json"
	"os"
)

func ReadFile[T any](path string, dest T) error {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Defer closing the file
	defer file.Close()

	// Read the file content
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&dest); err != nil {
		return err
	}

	// Return nil if no error occurred
	return nil
}
