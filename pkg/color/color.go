package color

import "fmt"

type ColorCode int

const (
	InfoColor    ColorCode = 34
	NoticeColor  ColorCode = 36
	WarningColor ColorCode = 33
	ErrorColor   ColorCode = 31
	DebugColor   ColorCode = 35
)

func Sprintf(str string, code ColorCode) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", code, str)
}
