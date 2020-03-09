package tests

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

type walkFunc func(string) error

// Walk walks through every code files in project.
func Walk(t *testing.T, baseDir string, excludes []string, wf walkFunc) {
	baseDir = path.Join("../", baseDir)

	filepath.Walk(baseDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(path, ".git") {
			return nil
		}

		if ext := filepath.Ext(path); ext != ".go" {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		for _, exclude := range excludes {
			p := strings.TrimPrefix(path, "../")
			if strings.HasPrefix(p, exclude) {
				return nil
			}
		}

		err = wf(path)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})
}

// ReadFile reads code file from disk.
func ReadFile(path string) []string {
	codeBytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	codes := strings.Split(string(codeBytes), "\n")
	return codes
}
