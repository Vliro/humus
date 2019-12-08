package parse

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	Parse(&Config{
		State:   "dgraph",
		Input:   "../../testing/",
		Output:  "../../testing/",
		Package: "gen",
	})
}

func TestDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		} else {
			if !strings.Contains(r.(error).Error(), "duplicate") {
				t.Fail()
			}
		}
	}()
	Parse(&Config{
		State:   "dgraph",
		Input:   "testdata/duplicate/",
		Output:  "/dev/null",
		Package: "gen",
	})
}
