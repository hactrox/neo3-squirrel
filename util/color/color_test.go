package color

import (
	"testing"
)

func TestRed(t *testing.T) {
	txt := Red("TEST")
	answer := "\033[0;31mTEST\033[0m"
	if txt != answer {
		t.Fatal("Failed render func 'Red'")
	}
}
