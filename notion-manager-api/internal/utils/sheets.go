package utils

import (
	"fmt"
	"regexp"
)

func ExtractSpreadsheetID(url string) (string, error) {
	re := regexp.MustCompile(`docs\.google\.com/spreadsheets/d/([a-zA-Z0-9-_]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Google Sheets URL")
	}
	return matches[1], nil
}
