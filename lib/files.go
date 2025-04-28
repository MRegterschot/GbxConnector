package lib

import (
	"encoding/json"
	"os"
)

func ReadFile[T any](path string, dest *T) error {
	// Open the file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the destination variable
	if err := json.Unmarshal(data, dest); err != nil {
		return err
	}

	// Return nil if no error occurred
	return nil
}
