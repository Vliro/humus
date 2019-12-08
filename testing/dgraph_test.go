package gen

import (
	"context"
	"fmt"
	"github.com/Vliro/humus"
	"github.com/dgraph-io/dgo/protos/api"
	"io/ioutil"
	"testing"
	"time"
)

var db *humus.DB

func TestMain(m *testing.M) {
	conf := &humus.Config{
		IP:         "localhost",
		Port:       9080,
		Tls:        false,
		RootCA:     "",
		NodeCRT:    "",
		NodeKey:    "",
		LogQueries: true,
	}
	db = humus.Init(conf, GetGlobalFields())
	dropAndSchema()
	addData()
	m.Run()
}

func dropAndSchema() {
	err := db.Alter(context.Background(), &api.Operation{
		DropAll: true,
	})
	if err != nil {
		panic(err)
	}
	sch, err := ioutil.ReadFile("schema.txt")
	if err != nil {
		panic(err)
	}
	err = db.Alter(context.Background(), &api.Operation{
		Schema: string(sch),
	})
	if err != nil {
		panic(err)
	}
}

//Example in adding data.
//This will add three nodes as the user node
//will be shared automatically.
func addData() {
	var q Question
	t := time.Now()
	q.Title = "First Question"
	q.DatePublished = &t
	var c = User{
		Name: "User",
	}
	q.From = &c
	var com = Comment{
		From: &c,
	}
	com.Text = "This is a comment"
	com.DatePublished = &t
	q.Comments = append(q.Comments, &com)
	mut := humus.CreateMutations(humus.MutateSet, &q)
	resp, err := db.Mutate(context.Background(), mut)
	if err != nil {
		panic(err)
	}
	if len(resp.Uids) != 3 {
		panic("data added is invalid, broken recurse")
	}
}

const upsertQuery = string(`query {
	user as var(func: eq(` + UserNameField + `,%s))
}
`)

//If user exists, set email to value.
func TestUpsert(t *testing.T) {
	var c User
	c.Email = "email@email.com"
	c.Uid = humus.UIDVariable("user")
	txn := db.NewTxn(false)
	defer txn.Discard(context.Background())
	mut := humus.CreateMutation(&c, humus.MutateSet)
	//Set the condition for update.
	mut.Condition = "@if(eq(len(user),1))"
	resp, err := txn.Upsert(context.Background(), humus.NewStaticQuery(fmt.Sprintf(upsertQuery, "User")), mut)
	if err != nil {
		t.Fail()
		return
	}
	if len(resp.Uids) > 0 || len(resp.Txn.Preds) != 1 {
		t.Fail()
		return
	}
	err = txn.Commit(context.Background())
	if err != nil {
		panic(err)
	}
	c = User{}
	err = db.Query(context.Background(), humus.GetByPredicate(UserNameField, UserFields, "User"), &c)
	if err != nil {
		t.Error(err)
		return
	}
	if c.Email != "email@email.com" {
		t.Fail()
		return
	}
}

func TestUpsertInsert(t *testing.T) {
	var c User
	c.Email = "email@email.com"
	c.Name = "User"
	txn := db.NewTxn(false)
	defer txn.Discard(context.Background())
	mut := humus.CreateMutation(&c, humus.MutateSet)
	//Set the condition for update.
	mut.Condition = "@if(eq(len(user),0))"
	resp, err := txn.Upsert(context.Background(), humus.NewStaticQuery(fmt.Sprintf(upsertQuery, "User")), mut)
	if err != nil {
		t.Fail()
		return
	}
	if len(resp.Uids) != 0{
		t.Fail()
		return
	}
}

//Example in getting a question. We only need the username field from the user so just select it.
var questionFields = QuestionFields.Sub(QuestionFromField, UserFields.Select(UserNameField)).
	Sub(QuestionCommentsField, CommentFields.
		Sub(CommentFromField, UserFields.Select(UserNameField)))

func TestGet(t *testing.T) {
	var q Question
	err := db.Query(context.Background(), humus.GetByPredicate(QuestionTitleField, questionFields, "First Question"), &q)
	if err != nil {
		t.Error(err)
		return
	}
	if q.From == nil || q.Comments == nil || q.Comments[0].From == nil || q.From.Uid != q.Comments[0].From.Uid {
		t.Fail()
		return
	}
}
