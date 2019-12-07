package parse

import (
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
