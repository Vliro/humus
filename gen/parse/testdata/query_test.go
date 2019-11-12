package mulgen

import (
	"context"
	"mulbase"
	"testing"
	"time"
)

var db *mulbase.DB
var uid string

func TestMain(m *testing.M) {
	var conf = mulbase.Config{
		IP:         "172.17.0.2",
		Port:       9080,
		Tls:        false,
		LogQueries: true,
	}
	db = mulbase.Init(&conf)
	if db == nil {
		return
	}
	db.SetSchema(GetGlobalFields())
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
	s := mulbase.SaveScalars(&c, txn)
	r,err := txn.RunQuery(context.Background(), s)
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
	if c.Uid != mulbase.UID(uid) {
		t.Error("could not find uid")
		return
	}
	txn.Discard(context.Background())
}

func TestMutate(t *testing.T) {
	txn := db.NewTxn(false)
	var c Comment
	c.DatePublished = time.Now()
	c.Text = "This is a trial run."
	s := mulbase.SaveScalars(&c, txn)
	_,err := txn.RunQuery(context.Background(), s)
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
	ch := txn.RunQueryAsync(context.Background(), s)
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