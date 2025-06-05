package structs

import (
	"strings"
)

type MessageFormat string

type ChatConfig struct {
	ManualRouting  bool           `json:"manualRouting"`
	MessageFormat MessageFormat `json:"overrideFormat"`
}

// Format the message according to the MessageFormat.
func (f *MessageFormat) FormatMessage(login string, nickName string, message string) string {
	msg := strings.ReplaceAll(f.String(), "{login}", login)
	msg = strings.ReplaceAll(msg, "{nickName}", nickName)
	msg = strings.ReplaceAll(msg, "{message}", message)
	return strings.TrimSpace(msg)
}

func (f *MessageFormat) String() string {
	if f == nil {
		return ""
	}
	return string(*f)
}
