package lib

import (
	"encoding/json"
	"fmt"
	"net"
)

func ConvertCallbackData(data any, target any) error {
	if err := json.Unmarshal([]byte(data.([]any)[0].(string)), target); err != nil {
		return fmt.Errorf("failed to unmarshal into target: %w", err)
	}

	return nil
}

func IsDockerInternalIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	_, dockerNet, _ := net.ParseCIDR("172.16.0.0/16") // default Docker bridge subnet
	return dockerNet.Contains(parsedIP)
}
