package gen

//Code generated by mulgen. DO NOT EDIT (i mean, it will all be lost in the void)

import (
	"context"
	"github.com/Vliro/humus"
	"strconv"
	"time"
)

var _ context.Context
var _ time.Time
var _ humus.Fields
var _ = strconv.IntSize

//easyjson:json
type Event struct {
	//This line declares basic properties for a database node.
	humus.Node
	//Regular fields
	Name        string  `json:"Event.name,omitempty"`
	Attending   []*User `json:"Event.attending,omitempty"`
	Prices      []int   `json:"Event.prices,omitempty"`
	Description string  `json:"Event.description,omitempty"`
}

var EventFields humus.FieldList = []humus.Field{MakeField("Event.name", 0), MakeField("Event.attending", 0|humus.MetaObject|humus.MetaList|humus.MetaReverse), MakeField("Event.prices", 0|humus.MetaList), MakeField("Event.description", 0)}

//Generating constant field values.
const (
	EventNameField        humus.Predicate = "Event.name"
	EventAttendingField   humus.Predicate = "Event.attending"
	EventPricesField      humus.Predicate = "Event.prices"
	EventDescriptionField humus.Predicate = "Event.description"
)

//SaveValues saves the node values that
//do not contain any references to other objects.
//Not needed for now.
/*
func (r *Event) SaveValues(ctx context.Context, txn *humus.Txn) error {
    mut := humus.CreateMutation(r.Values(), humus.MutateSet)
    _,err := txn.Mutate(ctx, mut)
    return err
}
*/
func (r *Event) GetType() []string {
	return _EventTypes
}

//Reset resets the node to only its UID
//component. Useful for saving to the database.
//Calling this function ensures that, at most,
//the uid and type is serialized.
func (r *Event) Reset() {
	if r != nil {
		*r = Event{Node: r.Node}
	}
}

func (r *Event) Facets() []string {
	return nil
}

//Fields returns all Scalar fields for this value.
func (r *Event) Fields() humus.FieldList {
	return EventFields
}

var _EventTypes = []string{
	"Event",
}

//Sets the types. This DOES NOT include interfaces!
//as they are set in dgraph already.
func (r *Event) SetType() {
	r.Node.Type = _EventTypes
}

//Values returns all the scalar values for this node.
func (r *Event) Values() humus.DNode {
	var m EventScalars
	m.Name = r.Name
	m.Prices = r.Prices
	m.Description = r.Description
	r.SetType()
	m.Node = r.Node
	return &m
}

/*
//Values returns all the scalar values for this node.
//Note that this completely ignores any omitempty tags so use with care.
func (r *Event) MapValues() humus.Mapper {
   var m = make(map[string]interface{})
   m["Event.name"]= r.Name
      m["Event.prices"]= r.Prices
      m["Event.description"]= r.Description
      if r.Uid != "" {
      m["uid"] = r.Uid
   }
    r.SetType()
    m["dgraph.type"] = r.typ
    return m
}
*/

//EventScalars is simply to avoid a map[string]interface{}
//It is a mirror of the previous struct with all scalar values.
type EventScalars struct {
	humus.Node
	Name        string `json:"Event.name,omitempty"`
	Prices      []int  `json:"Event.prices,omitempty"`
	Description string `json:"Event.description,omitempty"`
}

func (s *EventScalars) Values() humus.DNode {
	return s
}

func (s *EventScalars) Fields() humus.FieldList {
	return EventFields
}

//End of model.template
//easyjson:json
type User struct {
	//This line declares basic properties for a database node.
	humus.Node
	//Regular fields
	Name      string   `json:"User.name,omitempty"`
	Email     string   `json:"User.email,omitempty"`
	FullName  string   `json:"User.fullName,omitempty"`
	Attending []*Event `json:"~Event.attending,omitempty"`
	Premium   int      `json:"Event.attending|premium,omitempty"`
}

var UserFields humus.FieldList = []humus.Field{MakeField("User.name", 0), MakeField("User.email", 0), MakeField("User.fullName", 0), MakeField("~Event.attending", 0|humus.MetaObject|humus.MetaList|humus.MetaReverse), MakeField("Event.attending|premium", 0|humus.MetaFacet)}

//Generating constant field values.
const (
	UserNameField      humus.Predicate = "User.name"
	UserEmailField     humus.Predicate = "User.email"
	UserFullNameField  humus.Predicate = "User.fullName"
	UserAttendingField humus.Predicate = "~Event.attending"
	UserPremiumField   humus.Predicate = "Event.attending|premium"
)

//SaveValues saves the node values that
//do not contain any references to other objects.
//Not needed for now.
/*
func (r *User) SaveValues(ctx context.Context, txn *humus.Txn) error {
    mut := humus.CreateMutation(r.Values(), humus.MutateSet)
    _,err := txn.Mutate(ctx, mut)
    return err
}
*/
func (r *User) GetType() []string {
	return _UserTypes
}

//Reset resets the node to only its UID
//component. Useful for saving to the database.
//Calling this function ensures that, at most,
//the uid and type is serialized.
func (r *User) Reset() {
	if r != nil {
		*r = User{Node: r.Node}
	}
}

func (r *User) Facets() []string {
	return nil
}

//Fields returns all Scalar fields for this value.
func (r *User) Fields() humus.FieldList {
	return UserFields
}

var _UserTypes = []string{
	"User",
}

//Sets the types. This DOES NOT include interfaces!
//as they are set in dgraph already.
func (r *User) SetType() {
	r.Node.Type = _UserTypes
}

//Values returns all the scalar values for this node.
func (r *User) Values() humus.DNode {
	var m UserScalars
	m.Name = r.Name
	m.Email = r.Email
	m.FullName = r.FullName
	m.Premium = r.Premium
	r.SetType()
	m.Node = r.Node
	return &m
}

/*
//Values returns all the scalar values for this node.
//Note that this completely ignores any omitempty tags so use with care.
func (r *User) MapValues() humus.Mapper {
   var m = make(map[string]interface{})
   m["User.name"]= r.Name
      m["User.email"]= r.Email
      m["User.fullName"]= r.FullName
      m["Event.attending|premium"]= r.Premium
      if r.Uid != "" {
      m["uid"] = r.Uid
   }
    r.SetType()
    m["dgraph.type"] = r.typ
    return m
}
*/

//UserScalars is simply to avoid a map[string]interface{}
//It is a mirror of the previous struct with all scalar values.
type UserScalars struct {
	humus.Node
	Name     string `json:"User.name,omitempty"`
	Email    string `json:"User.email,omitempty"`
	FullName string `json:"User.fullName,omitempty"`
	Premium  int    `json:"Event.attending|premium,omitempty"`
}

func (s *UserScalars) Values() humus.DNode {
	return s
}

func (s *UserScalars) Fields() humus.FieldList {
	return UserFields
}

//End of model.template
//easyjson:json
type Error struct {
	//This line declares basic properties for a database node.
	humus.Node
	//Regular fields
	Message   string     `json:"Error.message,omitempty"`
	ErrorType string     `json:"Error.errorType,omitempty"`
	Time      *time.Time `json:"Error.time,omitempty"`
}

var ErrorFields humus.FieldList = []humus.Field{MakeField("Error.message", 0), MakeField("Error.errorType", 0), MakeField("Error.time", 0)}

//Generating constant field values.
const (
	ErrorMessageField   humus.Predicate = "Error.message"
	ErrorErrorTypeField humus.Predicate = "Error.errorType"
	ErrorTimeField      humus.Predicate = "Error.time"
)

//SaveValues saves the node values that
//do not contain any references to other objects.
//Not needed for now.
/*
func (r *Error) SaveValues(ctx context.Context, txn *humus.Txn) error {
    mut := humus.CreateMutation(r.Values(), humus.MutateSet)
    _,err := txn.Mutate(ctx, mut)
    return err
}
*/
func (r *Error) GetType() []string {
	return _ErrorTypes
}

//Reset resets the node to only its UID
//component. Useful for saving to the database.
//Calling this function ensures that, at most,
//the uid and type is serialized.
func (r *Error) Reset() {
	if r != nil {
		*r = Error{Node: r.Node}
	}
}

func (r *Error) Facets() []string {
	return nil
}

//Fields returns all Scalar fields for this value.
func (r *Error) Fields() humus.FieldList {
	return ErrorFields
}

var _ErrorTypes = []string{
	"Error",
}

//Sets the types. This DOES NOT include interfaces!
//as they are set in dgraph already.
func (r *Error) SetType() {
	r.Node.Type = _ErrorTypes
}

//Values returns all the scalar values for this node.
func (r *Error) Values() humus.DNode {
	var m ErrorScalars
	m.Message = r.Message
	m.ErrorType = r.ErrorType
	m.Time = r.Time
	r.SetType()
	m.Node = r.Node
	return &m
}

/*
//Values returns all the scalar values for this node.
//Note that this completely ignores any omitempty tags so use with care.
func (r *Error) MapValues() humus.Mapper {
   var m = make(map[string]interface{})
   m["Error.message"]= r.Message
      m["Error.errorType"]= r.ErrorType
      m["Error.time"]= r.Time
      if r.Uid != "" {
      m["uid"] = r.Uid
   }
    r.SetType()
    m["dgraph.type"] = r.typ
    return m
}
*/

//ErrorScalars is simply to avoid a map[string]interface{}
//It is a mirror of the previous struct with all scalar values.
type ErrorScalars struct {
	humus.Node
	Message   string     `json:"Error.message,omitempty"`
	ErrorType string     `json:"Error.errorType,omitempty"`
	Time      *time.Time `json:"Error.time,omitempty"`
}

func (s *ErrorScalars) Values() humus.DNode {
	return s
}

func (s *ErrorScalars) Fields() humus.FieldList {
	return ErrorFields
}

//End of model.template