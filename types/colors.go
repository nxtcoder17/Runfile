package types

import "fmt"

const (
	StyleReset   = "\033[0m"
	StyleBold    = "\033[1m"
	StyleFgGreen = "\033[32m"
	StyleFgRed   = "\033[31m"
	StyleFgGray  = "\033[90m"
	// StyleFgGray  = "\033[2;30m"
)

func GetStyledPrefix(prefix string) string {
	return fmt.Sprintf("%s[%s]%s ", StyleFgGreen, prefix, StyleReset)
	// return fmt.Sprintf("%s%s |%s ", Green, prefix, Reset)
}

func GetDimStyledPrefix(prefix string) string {
	return fmt.Sprintf("%s[%s]%s ", "\033[3;36m", prefix, StyleReset)
	// return fmt.Sprintf("%s%s |%s ", Green, prefix, Reset)
}

func GetDimmedText(text []byte) string {
	return fmt.Sprintf("%s%s%s", "\033[0;36m", text, StyleReset)
	// return fmt.Sprintf("%s%s%s", StyleFgGray, text, StyleReset)
}

func GetCommandHighlight(text []byte) string {
	return fmt.Sprintf("%s%s%s", StyleFgGreen, text, StyleReset)
}

func GetErrorStyledPrefix(prefix string) string {
	return fmt.Sprintf("%s[%s]%s ", StyleFgRed, prefix, StyleReset)
	// return fmt.Sprintf("%s%s |%s ", Green, prefix, Reset)
}
