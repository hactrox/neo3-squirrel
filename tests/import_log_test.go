package tests

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLogImport(t *testing.T) {
	excludes := []string{
		"config/",
		"tests/",
	}

	Walk(t, "", excludes, checkLogImport)
}

func checkLogImport(path string) error {
	codes := ReadFile(path)

	for line, code := range codes {
		code = strings.TrimSpace(code)
		code = strings.TrimPrefix(code, "import ")
		code = strings.Trim(code, "\"")

		if code == "log" {
			_, fullpath, _, _ := runtime.Caller(0)
			path = filepath.Join(filepath.Dir(fullpath), path)
			return fmt.Errorf("Use \"neo3-squirrel/log\" instead of \"log\" in\n%s:%d", path, line+1)
		}
	}

	return nil
}
