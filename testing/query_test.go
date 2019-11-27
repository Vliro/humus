package gen

import (
	"context"
	"fmt"
	"github.com/Vliro/mulbase"
	"github.com/dgraph-io/dgo/protos/api"
	"io/ioutil"
	"os"
	"testing"
	"time"
)
var db *mulbase.DB
//Setup schema and drop previous.
func TestMain(m *testing.M) {
	var conf = &mulbase.Config{
		IP:         "localhost",
		Port:       9080,
		Tls:        false,
		RootCA:     "",
		NodeCRT:    "",
		NodeKey:    "",
		LogQueries: true,
	}
	db = mulbase.Init(conf, GetGlobalFields())
	defer db.Cleanup(context.Background())
	var op = new(api.Operation)
	op.DropAll = true
	err := db.Alter(context.Background(), op)
	if err != nil {
		panic(err)
	}
	schema, err := ioutil.ReadFile("dgraph.txt")
	if err != nil {
		panic(err)
	}
	op.DropAll = false
	op.Schema = string(schema)
	err = db.Alter(context.Background(), op)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestMutate(t *testing.T) {
	var p Post
	p.Text = "Test"
	p.DatePublished = time.Now()
	val, err := db.Mutate(context.Background(), mulbase.SaveNode(&p))
	if err != nil {
		t.Error(err)
		return
	}
	if len(val.Uids) == 0 {
		t.Fail()
		return
	}
	p = Post{}
	var ui mulbase.UID
	for _,v := range val.Uids {
		ui = mulbase.UID(v)
	}
	err = db.Query(context.Background(), mulbase.GetByUid(ui, PostFields), &p)
	if p.Text != "Test" {
		t.Fail()
		return
	}
}

func TestEdge(t *testing.T) {
	tim := time.Now()
	var p Comment
	//Test enum.
	p.Test = EMPIRE
	p.CommentsOn = &Post{
		Text:          "swag",
		DatePublished: time.Now(),
	}
	p.Text = "Testing"
	p.DatePublished = time.Now()
	n, err := db.Mutate(context.Background(), mulbase.CreateMutation(&p, mulbase.QuerySet))
	if err != nil || len(n.Uids) != 2  {
		t.Fail()
		return
	}
	p = Comment{}
	var fields = CommentFields.Sub(CommentCommentsOnField, PostFields)
	err = db.Query(context.Background(), mulbase.NewQuery().SetFunction(mulbase.MakeFunction(mulbase.FunctionHas).AddPred(PostTextField)).SetFields(fields), &p)
	if err != nil || p.Uid == "" || p.CommentsOn == nil || p.CommentsOn.Uid == "" || p.CommentsOn.Text != "swag" || p.Test != EMPIRE {
		t.Fail()
		return
	}
	fmt.Printf("Time to execute TestEdge %v", time.Now().Sub(tim))
}

func TestQuery(t *testing.T) {
	var p Comment
	p.CommentsOn = &Post{
		Text:          "TestQuery",
		DatePublished: time.Now(),
	}
	p.Text = "TestQueryComment"
	p.DatePublished = time.Now()
	n, err := db.Mutate(context.Background(), mulbase.CreateMutation(&p, mulbase.QuerySet))
	if err != nil || len(n.Uids) != 2  {
		t.Fail()
		return
	}
	p = Comment{}
	err = db.Query(context.Background(),mulbase.GetByPredicate(PostTextField, CommentFields, "TestQueryComment"), &p)
	if p.Uid == "" || err != nil {
		t.Fail()
		return
	}
}


/*
var db *mulbase.DB
var uid string

func TestMain(m *testing.M) {
	var conf = mulbase.Config{
		IP:         "172.17.0.2",
		Port:       9080,
		Tls:        false,
		LogQueries: true,
	}
	db = mulbase.Init(&conf, GetGlobalFields())
	if db == nil {
		return
	}
	uidd := runOneMutation(db)
	if uidd == "" {
		return
	}
	uid = uidd
	m.Run()
}
//returns uid for one object.
func runOneMutation(d *mulbase.DB) string{
	txn := db.NewTxn(false)
	var c Comment
	c.DatePublished = time.Now()
	c.Text = "First"
	s := mulbase.SaveScalars(&c)
	err := txn.Query(context.Background(), s)
	if err != nil {
		return ""
	}
	err = txn.Commit(context.Background())
	if err != nil {
		return ""
	}
	for _,v := range r.Uids {
		return v
	}
	return ""
}

func TestHasUid(t *testing.T) {
	var c Comment
	txn := db.NewTxn(true)
	err := mulbase.GetByUid(context.Background(), uid, CommentFields, txn, &c)
	if err != nil {
		t.Error(err)
		return
	}
	if c.Uid == "" || c.Uid != mulbase.UID(uid) {
		t.Error("could not find uid")
		return
	}
	_ = txn.Discard(context.Background())
}

func TestMutate(t *testing.T) {
	txn := db.NewTxn(false)
	var c Comment
	c.DatePublished = time.Now()
	c.Text = "This is a trial run."
	s := mulbase.SaveScalars(&c, txn)
	err := txn.Query(context.Background(), s)
	if err != nil {
		t.Error(err)
		return
	}
	err = txn.Commit(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
}

func TestMutateAsync(t *testing.T) {
	txn := db.NewTxn(false)
	var c Comment
	c.DatePublished = time.Now()
	c.Text = "This is a trial run async."
	s := mulbase.SaveScalars(&c, txn)
	ch := txn.QueryAsync(context.Background(), s)
	r := <-ch
	if r.Err != nil {
		t.Error(r.Err)
		return
	}
	if len(r.Res.Uids) != 1 {
		t.Error("invalid UID count in TestMutateAsync")
		return
	}
	err := txn.Commit(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
}

func TestStaticQuery(t *testing.T) {
	q := mulbase.NewStaticQuery(fmt.Sprintf(`{q(func: uid(%s)) {
		uid
		Post.datePublished
		Post.text}}`, uid))
	txn := db.NewTxn(true)
	var c Comment
	_, err := txn.Query(context.Background(), q, &c)
	if err != nil {
		t.Error(err)
		return
	}
	if string(c.Uid) != uid {
		t.Error("could not fetch in static query.")
		return
	}
	_ = txn.Discard(context.Background())
}

 */