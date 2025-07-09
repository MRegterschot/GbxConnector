package lib

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/MRegterschot/GbxConnector/structs"
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

func FilterServersByUuid(allServers []structs.ServerResponse, subscriptionSet map[string]bool) []structs.ServerResponse {
	if len(subscriptionSet) == 0 {
		// No filter, return all
		return allServers
	}
	filtered := []structs.ServerResponse{}
	for _, server := range allServers {
		if subscriptionSet[server.Uuid] {
			filtered = append(filtered, server)
		}
	}
	return filtered
}
