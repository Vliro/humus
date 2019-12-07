package gen

import (
	"fmt"
	"github.com/Vliro/humus"
	"testing"
)

func TestReset (t *testing.T) {
	var a User
	a.Name = "Username"
	a.Uid = "0x1"
	a.Reset()
	if a.Name != "" || a.Uid != "0x1" {
		t.Fail()
		return
	}
}

func TestLang(t *testing.T) {
	const expected = "query t($0:string){q(func: eq(<Post.text>,$0)){Post.text@se:. Post.datePublished   uid}}"
	q := humus.NewQuery(PostFields)
	q.Language(humus.LanguageSwedish, false)
	q.Function(humus.Equals).PredValue(PostTextField, "Test")

	str, err := q.Process()
	if err != nil {
		t.Fail()
		return
	}
	fmt.Println(str)
}

func TestMutate(t *testing.T) {

}