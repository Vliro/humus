package mulbase

import (
	"bytes"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

//FieldList is a list of Fields associated with a generated object.
//These are of global type and should never be modified lest
//the state of the entire application should be changed.
//Whenever fields are added to this list they are copied.
type FieldList []Field

//No need to reallocate. Typing here is very important as it will no longer copy.
type NewList FieldList

//Sub allows you to add sub-field structures.
//TODO: Racy reads? There are never writes to a global field-list unless you are doing something wrong!
//but might be concurrent reads.
func (f FieldList) Sub(name string, fl []Field) NewList {
	var newArr NewList = make([]Field, len(f))
	//Copy!
	copy(newArr, f)
	//linear search but there are not a lot of values. Hash-map feels overkill
	for k, v := range newArr {
		if v.Name == name {
			var newField = Field{
				Name:   v.Name,
				Fields: fl,
			}
			newArr[k] = newField
			break
		}
	}
	return newArr
}

func (f NewList) Sub(name string, fl []Field) NewList {
	//linear search but fast either way.
	for k, v := range f {
		if v.Name == name {
			f[k].Fields = fl
			break
		}
	}
	return f
}

func (f NewList) Add(fi Field) NewList {
	return append(f, fi)
}

//TODO: offset
type CountType string

const (
	CountFirst  CountType = "first"
	CountOffset CountType = "offset"
	CountAfter  CountType = "after"
)

type AggregateType string

const (
	TypeCount AggregateType = "count"
	TypeSum   AggregateType = "sum"
	TypeVar   AggregateType = "var"
	TypeVal   AggregateType = "val"
)

type Count struct {
	Type  CountType
	Value int
}

//A meta field for schemas.
type FieldMeta int

func (f FieldMeta) Lang() bool {
	return f&MetaLang > 0
}

func (f FieldMeta) List() bool {
	return f&MetaLang > 0
}

func (f FieldMeta) Reverse() bool {
	return f&MetaReverse > 0
}

const (
	MetaObject FieldMeta = 1 << iota
	MetaList
	MetaLang
	MetaUid
	MetaReverse
)

// Field is a recursive data struct which represents a GraphQL query field.
type Field struct {
	Name   string
	Fields []Field
	Meta   FieldMeta
	Local bool
	//Type VarType
}

// MakeField constructs a Field of given name and returns the Field.
func MakeField(name string, meta FieldMeta) Field {
	//TODO: better facet support
	var x = Field{Name: name, Meta: meta}
	return x
}

// String returns read only create token channel or an error.
// It checks if there is a circle.
func (f *Field) String(q *GeneratedQuery, parent string, sb *bytes.Buffer) error {
	if err := f.check(q); err != nil {
		// return a closed channel instead of nil for receiving from nil blocks forever, hard to debug and confusing to users.
		return errors.WithStack(err)
	}
	f.string(q, parent, sb)
	return nil
}

func (f *Field) writeLanguageTag(sb *bytes.Buffer, l Language) {
	if l != LanguageDefault && l != LanguageNone {
		sb.WriteString("@" + string(l) + ":.")
	}
}

func (f *Field) getName() string {
	if f.Name[0] == '~' {
		return f.Name[1:]
	}
	return f.Name
}

// One may have noticed that there is a public String and a private create.
// The different being the public method checks the validity of the Field structure
// while the private counterpart assumes the validity.
// Returns whether this field is a facet field.
func (f *Field) string(q *GeneratedQuery, parent string, sb *bytes.Buffer) bool {
	agg, ok := q.FieldAggregate[parent]
	if ok {
		sb.WriteString(agg.Alias)
		sb.WriteString(" : ")
		sb.WriteString(string(agg.Type))
		sb.WriteString(tokenLP)
		sb.WriteString(f.Name)
		//if f.SchemaField.Lang {
		//	f.writeLanguageTag(sb, q.Language)
		//}
		sb.WriteString(tokenRP)
	} else {
		sb.WriteString(f.Name)
		//if f.SchemaField.Lang {
		//	f.writeLanguageTag(sb, q.Language)
		//}
	}

	//Handle the (first: -1)
	val, ok := q.FieldCount[parent]
	if ok {
		sb.WriteString(tokenLP)
		for k, v := range val {
			if k != 0 {
				sb.WriteString(tokenComma)
			}
			sb.WriteString(string(v.Type))
			sb.WriteString(tokenColumn)
			sb.WriteString(strconv.Itoa(v.Value))
		}
		sb.WriteString(tokenRP)
	}
	order, ok := q.FieldOrderings[parent]
	if ok {
		sb.WriteString(tokenLP)
		for k, v := range order {
			sb.WriteString(v.String())
			if k != 0 {
				sb.WriteString(tokenComma)
			}
		}
		sb.WriteString(tokenRP)
	}
	filter, ok := q.FieldFilters[parent]
	if ok {
		sb.WriteString(tokenSpace)
		filter.create(q, parent, sb)
	}
	var tmp = new(bytes.Buffer)
	//writefct := false
	if len(f.Fields) > 0 {
		tmp.WriteString(tokenLB)
		for i, field := range f.Fields {
			if len(field.Name) > 0 {
				if i != 0 && i != len(f.Fields) {
					tmp.WriteString(tokenSpace)
				}
				field.string(q, parent+"/"+field.Name, tmp)
			}
			/*if field.Facet && !writefct {
				sb.WriteString(tokenSpace)
				sb.WriteString("@facets")
				sb.WriteString(tokenSpace)
				writefct = true
			}*/
		}
		tmp.WriteString(tokenSpace)
		tmp.WriteString("uid")
		tmp.WriteString(tokenSpace)
		tmp.WriteString(tokenRB)
	}
	io.Copy(sb, tmp)
	return false
}

func (f *Field) check(q *GeneratedQuery) error {
	return nil
}

func MakeFields(s ...string) []Field {
	var f []Field
	f = make([]Field, len(s))
	for k, v := range s {
		ft := MakeField(v, 0)
		f[k] = ft
	}
	return f
}
