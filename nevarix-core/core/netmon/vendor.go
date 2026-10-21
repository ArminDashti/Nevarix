package netmon

import "strings"

func IdentifyVendor(sysDescr, sysObjectID string) string {
	lower := strings.ToLower(sysDescr + " " + sysObjectID)
	switch {
	case strings.Contains(lower, "mikrotik") || strings.Contains(sysObjectID, "1.3.6.1.4.1.14988"):
		return "mikrotik"
	case strings.Contains(lower, "cisco") || strings.Contains(sysObjectID, "1.3.6.1.4.1.9"):
		return "cisco"
	default:
		return "unknown"
	}
}
