package gen

import (
	"testing"
)

func TestRelative(t *testing.T) {
	var a = new(Post)
	a.Text = "Value"
	a.Uid = "0x1"
	a = a.Relative()
	if a.Text != "" || a.Uid != "0x1" {
		t.Fail()
		return
	}
}

func TestRecurse(t *testing.T) {
	var q Question
	q.Comments = append(q.Comments, &Comment{
		From: &User{Name: "User"},
		Post: Post{
			Text: "Text",
		},
	})
	q.Recurse(0)
	if q.Uid != "_:0" || q.Comments[0].Uid != "_:1" || q.Comments[0].From.Uid != "_:2" {
		t.Fail()
		return
	}
}
