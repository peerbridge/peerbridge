package color

import "fmt"

type ColorCode int

const (
	Error   ColorCode = 31
	Success ColorCode = 32
	Warning ColorCode = 33
	Info    ColorCode = 34
	Debug   ColorCode = 35
	Notice  ColorCode = 36
	Reset   ColorCode = 0
)

func Sprintf(a interface{}, code ColorCode) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", code, a)
}

func WithBackground(a interface{}, code ColorCode) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", code+10, a)
}

func SprintfInt(a interface{}, num uint8) string {
	return fmt.Sprintf("\u001b[48;5;%dm %s \u001b[0m", num, a)
}
