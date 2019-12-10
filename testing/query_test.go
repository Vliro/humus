package gen

import (
	"context"
	"github.com/Vliro/humus"
	"testing"
)

func TestGroupBy(t *testing.T) {
	q := humus.NewQuery(questionFields).Function(humus.Equals).PredValue(QuestionTitleField, "test")
	q.Path(QuestionFromField).GroupBy(UserNameField).
		Variable("", "count(uid)", false)
	//Should be valid syntax, just give no answer.
	err := db.Query(context.Background(), q, nil)
	if err != nil {
		t.Fail()
		return
	}
}

