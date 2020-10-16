package color

import (
	"fmt"
	"strconv"
)

const (
	// common
	reset  = "\033[0m" // auto reset the rest of text to default color
	normal = 0
	bold   = 1 // increase this value if you want bolder text
	// special
	dim       = 2
	underline = 4
	blink     = 5
	reverse   = 7
	hidden    = 8
	// color
	black       = 30 // default = 39
	red         = 31
	green       = 32
	yellow      = 33
	blue        = 34
	purple      = 35 // purple = magenta
	cyan        = 36
	lightGray   = 37
	darkGray    = 90
	lightRed    = 91
	lightGreen  = 92
	lightYellow = 93
	lightBlue   = 94
	lightPurple = 95
	lightCyan   = 96
	white       = 97
)

// Render rends text with parameters
func Render(colorCode int, fontSize int, content string) string {
	return "\033[" + strconv.Itoa(fontSize) + ";" + strconv.Itoa(colorCode) + "m" + content + reset
}

// Black text (use this with caution since most geeks use dark console)
func Black(txt string) string {
	return Render(black, normal, txt)
}

// Red text
func Red(txt string) string {
	return Render(red, normal, txt)
}

// Redf text
func Redf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return Red(msg)
}

// Green text
func Green(txt string) string {
	return Render(green, normal, txt)
}

// Greenf text
func Greenf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return Green(msg)
}

// Yellow text
func Yellow(txt string) string {
	return Render(yellow, normal, txt)
}

// Yellowf text
func Yellowf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return Yellow(msg)
}

// Blue text
func Blue(txt string) string {
	return Render(blue, normal, txt)
}

// Purple text
func Purple(txt string) string {
	return Render(purple, normal, txt)
}

// Cyan text
func Cyan(txt string) string {
	return Render(cyan, normal, txt)
}

// LightGray text
func LightGray(txt string) string {
	return Render(lightGray, normal, txt)
}

// DarkGray text
func DarkGray(txt string) string {
	return Render(darkGray, normal, txt)
}

// LightRed text
func LightRed(txt string) string {
	return Render(lightRed, normal, txt)
}

// LightGreen text
func LightGreen(txt string) string {
	return Render(lightGreen, normal, txt)
}

// LightYellow text
func LightYellow(txt string) string {
	return Render(lightYellow, normal, txt)
}

// LightBlue text
func LightBlue(txt string) string {
	return Render(lightBlue, normal, txt)
}

// LightPurple text
func LightPurple(txt string) string {
	return Render(lightPurple, normal, txt)
}

// LightCyan text
func LightCyan(txt string) string {
	return Render(lightCyan, normal, txt)
}

// LightCyanf text
func LightCyanf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return LightCyan(msg)
}

// White text
func White(txt string) string {
	return Render(white, normal, txt)
}

// BRed returns bold red test
func BRed(txt string) string {
	return Render(red, bold, txt)
}

// BRedf returns bold red test
func BRedf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return BRed(msg)
}

// BGreen returns bold green
func BGreen(txt string) string {
	return Render(green, bold, txt)
}

// BGreenf returns bold green
func BGreenf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return BGreen(msg)
}

// BYellow returns bold yellow
func BYellow(txt string) string {
	return Render(yellow, bold, txt)
}

// BYellowf returns bold yellow
func BYellowf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return BYellow(msg)
}

// BBlue returns bold blue
func BBlue(txt string) string {
	return Render(blue, bold, txt)
}

// BPurple returns bold purple
func BPurple(txt string) string {
	return Render(purple, bold, txt)
}

// BPurplef returns bold purple
func BPurplef(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return BPurple(msg)
}

// BCyan returns bold cyan
func BCyan(txt string) string {
	return Render(cyan, bold, txt)
}

// BLightGray returns bold light gray
func BLightGray(txt string) string {
	return Render(lightGray, bold, txt)
}

// BDarkGray returns bold dark gray
func BDarkGray(txt string) string {
	return Render(darkGray, bold, txt)
}

// BLightRed returns bold light red
func BLightRed(txt string) string {
	return Render(lightRed, bold, txt)
}

// BLightGreen returns bold light green
func BLightGreen(txt string) string {
	return Render(lightGreen, bold, txt)
}

// BLightGreenf returns bold light green
func BLightGreenf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return BLightGreen(msg)
}

// BLightYellow returns bold light yellow
func BLightYellow(txt string) string {
	return Render(lightYellow, bold, txt)
}

// BLightBlue returns bold light blue
func BLightBlue(txt string) string {
	return Render(lightBlue, bold, txt)
}

// BLightPurple returns bold light purple
func BLightPurple(txt string) string {
	return Render(lightPurple, bold, txt)
}

// BLightCyan returns bold light cyan
func BLightCyan(txt string) string {
	return Render(lightCyan, bold, txt)
}

// BLightCyanf returns bold light cyan
func BLightCyanf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	return BLightCyan(msg)
}

// BWhite returns bold white
func BWhite(txt string) string {
	return Render(white, bold, txt)
}

// GRed returns text with red background
func GRed(txt string) string {
	return Render(red+1, normal, txt)
}

// GGreen returns text with green background
func GGreen(txt string) string {
	return Render(green+1, normal, txt)
}

// GYellow returns text with yellow background
func GYellow(txt string) string {
	return Render(yellow+1, normal, txt)
}

// GBlue returns text with blue background
func GBlue(txt string) string {
	return Render(blue+1, normal, txt)
}

// GPurple returns text with purple background
func GPurple(txt string) string {
	return Render(purple+1, normal, txt)
}

// GCyan returns text with cyan background
func GCyan(txt string) string {
	return Render(cyan+1, normal, txt)
}

// GLightGray returns text with light gray background
func GLightGray(txt string) string {
	return Render(lightGray+1, normal, txt)
}

// GDarkGray returns text with dark gray background
func GDarkGray(txt string) string {
	return Render(darkGray+1, normal, txt)
}

// GLightRed returns text with light red background
func GLightRed(txt string) string {
	return Render(lightRed+1, normal, txt)
}

// GLightGreen returns text with light green background
func GLightGreen(txt string) string {
	return Render(lightGreen+1, normal, txt)
}

// GLightYellow returns text with light yellow background
func GLightYellow(txt string) string {
	return Render(lightYellow+1, normal, txt)
}

// GLightBlue returns text with blue background
func GLightBlue(txt string) string {
	return Render(lightBlue+1, normal, txt)
}

// GLightPurple returns text with light purple background
func GLightPurple(txt string) string {
	return Render(lightPurple+1, normal, txt)
}

// GLightCyan returns text with light cyan background
func GLightCyan(txt string) string {
	return Render(lightCyan+1, normal, txt)
}

// GWhite returns text with give text a white background
func GWhite(txt string) string {
	return Render(white+1, normal, txt)
}

// Bold returns bold text
func Bold(txt string) string {
	return Render(bold, normal, txt)
}

// Dim returns dimmed text
func Dim(txt string) string {
	return Render(dim, normal, txt)
}

// Underline returns underlined text
func Underline(txt string) string {
	return Render(underline, 0, txt)
}

// Hide given text, useful for password input
func Hide(txt string) string {
	return Render(hidden, normal, txt)
}
