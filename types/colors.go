package types

import "fmt"

const (
	StyleReset   = "\033[0m"
	StyleBold    = "\033[1m"
	StyleFgGreen = "\033[32m"
	StyleFgRed   = "\033[31m"
)

func GetStyledPrefix(prefix string) string {
	return fmt.Sprintf("%s[%s]%s ", StyleFgGreen, prefix, StyleReset)
	// return fmt.Sprintf("%s%s |%s ", Green, prefix, Reset)
}

func GetErrorStyledPrefix(prefix string) string {
	return fmt.Sprintf("%s[%s]%s ", StyleFgRed, prefix, StyleReset)
	// return fmt.Sprintf("%s%s |%s ", Green, prefix, Reset)
}
