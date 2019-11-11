package mulgen

import (
	"context"
	"fmt"
	"mulbase"
	"testing"
)

func TestQuery(t *testing.T) {
	db := mulbase.Init("172.17.0.2", 9080, false)
	if db == nil {
		t.Fail()
		return
	}
	db.SetSchema(GetGlobalFields())
	var q = mulbase.NewQuery()
	txn := db.NewTxn(true)
	txn.SetSchema(GetGlobalFields())
	q.Fields = CharacterFields.Sub("Character.name", TodoFields.Sub("Character.name", CharacterFields))
	q.Function = mulbase.MakeFunction(mulbase.FunctionEquals).AddPredValue("Character.name", "First", mulbase.TypeStr)
	var x Character
	_ = txn.RunQuery(context.Background(), q, &x)
	fmt.Println(x)
}


func TestAdd(t *testing.T) {
	db := mulbase.Init("172.17.0.2", 9080, false)
	if db == nil {
		t.Fail()
		return
	}
	db.SetSchema(GetGlobalFields())
	var q = mulbase.NewQuery()
	txn := db.NewTxn(false)
	txn.SetSchema(GetGlobalFields())
	var x Character
	x.Uid = "0x2"
	x.Name = "First"
	err := mulbase.AttachToListObject(&x, globalFields["Character.appearsIn"], txn, &Episode{
		Name: "Cool Episode",
	})
	_ = txn.Commit(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	_ = txn.RunQuery(context.Background(), q, &x)
	fmt.Println(x)
}

func TestScalars(t *testing.T) {
	db := mulbase.Init("172.17.0.2", 9080, false)
	if db == nil {
		t.Fail()
		return
	}
	db.SetSchema(GetGlobalFields())
	txn := db.NewTxn(false)
	var x Character
	x.Uid = "0x2"
	x.Name = "NewValue"
	err := mulbase.SaveScalars(&x, txn)

	err = txn.Commit(context.Background())
	txn = db.NewTxn(true)
	x = Character{}
	err = mulbase.GetByUid("0x2", CharacterFields, txn, &x)

	if err != nil {
		panic(err)
	}
}