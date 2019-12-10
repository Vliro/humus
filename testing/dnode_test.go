package gen

import (
	"testing"
)

func TestReset (t *testing.T) {
	var a Post
	a.Text = "Value"
	a.Uid = "0x1"
	a.Reset()
	if a.Text != "" || a.Uid != "0x1" {
		t.Fail()
		return
	}
}