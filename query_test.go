package mulbase

import (
	"context"
	"testing"
)

func TestQuery(t *testing.T) {
	q := GeneratedQuery{}
	txn := new(Txn)
	q.Fields = CharacterFields
	q.Function = MakeFunction(FunctionHas).AddPred("test")
	txn.RunQuery(context.Background(), &q, nil)
}

var CharacterFields FieldList = []Field{MakeField("Character.name"), MakeField("Character.appearsIn")}