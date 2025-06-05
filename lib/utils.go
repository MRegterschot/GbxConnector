package lib

import (
	"encoding/json"
	"fmt"
)

func ConvertCallbackData(data any, target any) error {
	if err := json.Unmarshal([]byte(data.([]any)[0].(string)), target); err != nil {
		return fmt.Errorf("failed to unmarshal into target: %w", err)
	}

	return nil
}
