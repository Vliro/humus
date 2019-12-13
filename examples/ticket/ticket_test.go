package gen

import (
	"context"
	"fmt"
	"github.com/Vliro/humus"
	"github.com/dgraph-io/dgo/protos/api"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestTicketQuery(t *testing.T) {
	ev, err := NewEvent("Event", "Best event", []int{1, 3, 5})
	if err != nil {
		t.Fail()
		return
	}
	var q = humus.NewQuery(eventFields).Function(humus.FunctionUid).Values(ev.Uid)

	q.At("", func(m humus.Mod) {
		m.Sort(humus.Descending, EventNameField)
	})

	str, _ := q.Process()
	fmt.Println(str)
}

func TestTicket(t *testing.T) {
	ev, err := NewEvent("Event", "Best event", []int{1, 3, 5})
	if err != nil {
		t.Fail()
		return
	}
	us, err := NewUser("User")
	if err != nil {
		t.Fail()
		return
	}
	uss, err := NewUser("Another user")
	if err != nil {
		t.Fail()
		return
	}
	ok := AddUserToEvent(ev.Uid, us.Uid, 1)
	ok = AddUserToEvent(ev.Uid, uss.Uid, 2)
	if !ok {
		t.Fail()
		return
	}
	//Returns an event with attending users ordered by the premium facet.
	ev, err = EventOrderByPremium(ev.Uid)
	if err != nil {
		t.Fail()
		return
	}
}

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
	makethousand()
	q := humus.NewQuery(eventFields).Function(humus.Type).Values("Event")
	var ev []*Event
	err = db.Query(context.Background(), q, &ev)
	if err != nil {
		panic(err)
	}
	us, err := GetUserWithAttending("User")
	_ = us
	m.Run()
}

func TestFacets(t *testing.T) {
	var q = humus.NewQuery(eventFields).Function(humus.Type).Values("Event")
	q.Facets(EventAttendingField, func(m humus.Mod) {
		m.Variable("result", "premium", false)
		m.Sort(humus.Descending, "premium")
	})
	q.At("", func(m humus.Mod) {
		m.Aggregate(humus.Sum, "result", "sum")
		m.Sort(humus.Descending, EventNameField)
		m.Paginate(humus.CountFirst, 5)
	})
	str, _ := q.Process()
	fmt.Println(str)
}

func makethousand() {
	var us = User{
		Name:     "User",
		Email:    "mail@mail.com",
		FullName: "User Man",
		Premium:  5,
	}
	var arr = make([]humus.DNode, 100)
	for i := 0; i < 100; i++ {
		arr[i] = &Event{
			Name:        "Event " + strconv.Itoa(i),
			Attending:   []*User{&us},
			Prices:      []int{1, 4, 1},
			Description: "Event description",
		}
	}
	mu := humus.CreateMutations(humus.MutateSet, arr...)
	resp, err := db.Mutate(context.Background(), mu)

	if err != nil {
		panic(err)
	}
	fmt.Println(len(resp.Uids), " nodes created.")
}
