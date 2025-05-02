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

func WriteFile[T any](path string, data *T) error {
	// Marshal the data into JSON format
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the file
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return err
	}

	return nil
}
