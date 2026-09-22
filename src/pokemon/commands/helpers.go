package commands

import "strings"

// extractKontoFromFilename extracts account name from ODS filename.
// Example: "Pokemon-GO-Adzik.ods" -> "Adzik"
func extractKontoFromFilename(filename string) string {
	name := strings.TrimSuffix(filename, ".ods")
	name = strings.TrimSuffix(name, ".ODS")
	const prefix = "Pokemon-GO-"
	if strings.HasPrefix(name, prefix) {
		return name[len(prefix):]
	}
	return ""
}
