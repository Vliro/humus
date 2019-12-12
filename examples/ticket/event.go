package gen

import (
	"context"
	"github.com/Vliro/humus"
)

var db *humus.DB

var eventFields = EventFields.Sub(EventAttendingField, UserFields)

func GetEvent(uid humus.UID) ([]*Event, error) {
	var ev []*Event
	err := db.Query(context.Background(), humus.GetByUid(uid, eventFields), &ev)
	return ev, err
}

//Adds a user to an event from a http endpoint. Premium facet.
func AddUserToEvent(euid humus.UID, userUid humus.UID, premium int) bool {
	var us User
	us.SetUID(userUid)
	us.Premium = premium
	var ev = Event{
		Attending: []*User{&us},
	}
	ev.SetUID(euid)
	_, err := db.Mutate(context.Background(), humus.CreateMutation(&ev, humus.MutateSet))
	if err != nil {
		return false
	}
	return true
}

func NewUser(name string) (*User, error) {
	var us = User{
		Name: name,
	}
	resp, err := db.Mutate(context.Background(), humus.CreateMutation(&us, humus.MutateSet))
	if err != nil {
		return nil, err
	}
	//One uid
	for _, v := range resp.Uids {
		us.Uid = humus.UID(v)
	}
	return &us, nil
}

func NewEvent(eventName string, description string, prices []int) (*Event, error) {
	var ev = Event{
		Name:        eventName,
		Prices:      prices,
		Description: description,
	}
	resp, err := db.Mutate(context.Background(), humus.CreateMutation(&ev, humus.MutateSet))
	if err != nil {
		return nil, err
	}
	//One uid
	for _, v := range resp.Uids {
		ev.Uid = humus.UID(v)
	}
	return &ev, nil
}

func EventOrderByPremium(uid humus.UID) (*Event, error) {
	var res Event
	var q = humus.NewQuery(eventFields).Function(humus.FunctionUid).Values(uid)
	q.Facets(EventAttendingField, func(m humus.Mod) {
		//Sort on the "premium" facet.
		m.Sort(humus.Descending, "premium")
		//We want the facet as well.
		m.Variable("", "premium", false)
	})
	err := db.Query(context.Background(), q, &res)
	return &res, err
}

func AttendingEvent(uid humus.UID) (*Event, int) {
	var res = struct {
		Event
		Value int `json:"result"`
	}{}
	var q = humus.NewQuery(EventFields).Function(humus.FunctionUid).Values(uid)
	q.At("", func(m humus.Mod) {
		m.Aggregate(humus.Count, string(EventAttendingField), "result")
	})
	err := db.Query(context.Background(), q, &res)
	if err != nil {
		return nil, -1
	}
	return &res.Event, res.Value
}
