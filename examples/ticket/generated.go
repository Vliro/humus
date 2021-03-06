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

func (r *Event) recurse(counter int) int {
	if r != nil {
		if r.Uid == "" {
			r.SetType()
			uid := humus.UID("_:" + strconv.Itoa(counter))
			r.Uid = uid
			counter++
		}
	} else {
		return counter
	}
	for _, v := range r.Attending {
		//no need to nil check as it is done in the function.
		counter = v.recurse(counter)
	}
	return counter
}

//Recurse iterates through the node and allocates type and UID
//to pointer nodes.
func (r *Event) Recurse(counter int) int {
	return r.recurse(counter)
}
func (r *User) recurse(counter int) int {
	if r != nil {
		if r.Uid == "" {
			r.SetType()
			uid := humus.UID("_:" + strconv.Itoa(counter))
			r.Uid = uid
			counter++
		}
	} else {
		return counter
	}
	for _, v := range r.Attending {
		//no need to nil check as it is done in the function.
		counter = v.recurse(counter)
	}
	return counter
}

//Recurse iterates through the node and allocates type and UID
//to pointer nodes.
func (r *User) Recurse(counter int) int {
	return r.recurse(counter)
}
func (r *Error) recurse(counter int) int {
	if r != nil {
		if r.Uid == "" {
			r.SetType()
			uid := humus.UID("_:" + strconv.Itoa(counter))
			r.Uid = uid
			counter++
		}
	} else {
		return counter
	}
	return counter
}

//Recurse iterates through the node and allocates type and UID
//to pointer nodes.
func (r *Error) Recurse(counter int) int {
	return r.recurse(counter)
}

//Beginning of field.template. General functions.
var globalFields = make(map[humus.Predicate]humus.Field)

func GetField(name humus.Predicate) humus.Field {
	return globalFields[name]
}

func MakeField(name humus.Predicate, flags humus.FieldMeta) humus.Field {
	var fi = humus.Field{Name: name, Meta: flags}
	globalFields[name] = fi
	return fi
}

func GetGlobalFields() map[humus.Predicate]humus.Field {
	return globalFields
}

type queryError struct {
	Msg string
}

func (q queryError) Error() string {
	return q.Msg
}

func newError(msg string) error {
	return queryError{msg}
}
