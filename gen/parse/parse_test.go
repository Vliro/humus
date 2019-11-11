package parse

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	Parse("testdata", "/tmp")
	time.Sleep(time.Second)
}
