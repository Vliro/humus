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
/*
func BenchmarkFast(b *testing.B) {
	var c Comment
	c.Text = "swag"

	c.From = &User{Name: "yolo"}

	byt, _ := easyjson.Marshal(&c)

	for i := 0; i < b.N; i++ {
		easyjson.Unmarshal(byt, &c)
	}
	easyjson.
	b.ReportAllocs()
}

func BenchmarkSlow(b *testing.B) {
	var c Comment
	c.Text = "swag"

	c.From = &User{Name: "yolo"}

	for i := 0; i < b.N; i++ {
		json.Marshal(&c)
	}
	b.ReportAllocs()
}

var iter = jsoniter.ConfigCompatibleWithStandardLibrary

func BenchmarkJsoniter(b *testing.B) {
	var c Comment
	c.Text = "swag"

	c.From = &User{Name: "yolo"}

	for i := 0; i < b.N; i++ {
		iter.Marshal(&c)
	}
	b.ReportAllocs()
}*/
