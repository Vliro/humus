package gen

import (
	"fmt"
	"github.com/Vliro/humus"
	"testing"
	"time"
)

//Parse test is in the testing folder for a simple reason, access the models.

const res = `{"q0":[{"Question.title":"First Question","Question.from":{"User.name":"User","uid":"0x75da"},"Question.comments":[{"Comment.from":{"User.name":"User","uid":"0x75da"},"Post.text":"This is a comment","Post.datePublished":"2019-12-09T00:06:16.254180118+01:00","uid":"0x75db"}],"Post.datePublished":"2019-12-09T00:06:16.254180118+01:00","uid":"0x75d9"}],"q1":[{"Question.title":"First Question","Question.from":{"User.name":"User","uid":"0x75da"},"Question.comments":[{"Comment.from":{"User.name":"User","uid":"0x75da"},"Post.text":"This is a comment","Post.datePublished":"2019-12-09T00:06:16.254180118+01:00","uid":"0x75db"}],"Post.datePublished":"2019-12-09T00:06:16.254180118+01:00","uid":"0x75d9"}]}`
func TestDeserialize(t *testing.T) {
	tt := time.Now()
	var q, q1 Question
	arr := append([]interface{}{}, &q, &q1)
	err := humus.HandleResponseFast([]byte(res), arr, []string{"q0", "q1"})
	if err != nil {
		t.Fail()
		return
	}
	fmt.Println(time.Now().Sub(tt))
}