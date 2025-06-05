package structs

import (
	"strings"
)

type OverrideFormat string

type ChatConfig struct {
	ManualRouting  bool           `json:"manualRouting"`
	OverrideFormat OverrideFormat `json:"overrideFormat"`
}

// Format the message according to the OverrideFormat.
func (f *OverrideFormat) FormatMessage(login string, nickName string, message string) string {
	msg := strings.ReplaceAll(f.String(), "{login}", login)
	msg = strings.ReplaceAll(msg, "{nickName}", nickName)
	msg = strings.ReplaceAll(msg, "{message}", message)
	return strings.TrimSpace(msg)
}

func (f *OverrideFormat) String() string {
	if f == nil {
		return ""
	}
	return string(*f)
}
