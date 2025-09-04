package tg

import "regexp"

// escapeMarkdownV2 escapes special characters for Telegram MarkdownV2
func EscapeMarkdownV2(text string) string {
	// Characters that need to be escaped in MarkdownV2: _*[]()~`>#+-=|{}.!
	specialChars := regexp.MustCompile(`([_*\[\]()~` + "`" + `>#+=|{}.!-])`)
	return specialChars.ReplaceAllString(text, "\\$1")
}
