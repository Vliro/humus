package humus

import (
	"errors"
	"strconv"
	"strings"
)

/*
	FieldList is a list of fields associated with a generated object.
	These are of global type and should never be modifierType lest
	the state of the entire application should be changed.
	Whenever fields are added to this list they are copied.
	Example usage:
	var NewFields = CharacterFields.Sub("Character.friends", CharacterFields).Sub("Character.enemies", CharacterFields.
					Sub("Character.items", ItemFields)
	This will also ensure fields are copied properly from the global list.
*/
//Fields is an interface for all possible type of fields.  This includes global fields as well as
//manually generated fields.
type Fields interface {
	//Sub allows you to create a sublist of predicates.
	//If there is an edge on a predicate name, then subbing on that
	//predicate gets all fields as specified by the fields interfaces.
	Sub(name Predicate, fields Fields) Fields
	//Add a field to this list.
	Add(fi Field) Fields
	//Get the fields as a slice
	Get() []Field
	//Len is the length of the fields.
	Len() int
}

//Select selects a subset of fields and returns a new list
//keeping all valid meta.
func Select(sch SchemaList, names ...Predicate) Fields {
	var fields = make(NewList, len(names))
	for k, v := range names {
		fields[k] = sch[v]
	}
	return fields
}

var emptyList FieldList = []Field{{
	Meta:   MetaEmpty,
	Name:   "",
	Fields: nil,
}}

type FieldList []Field

func (f FieldList) Get() []Field {
	return f
}

//Select allows you to perform selection of fields early, at init.
//This removes the Ignore meta from any field selected to allow password fields.
func (f FieldList) Select(names ...Predicate) Fields {
	var newList = make(NewList, len(names))
	index := 0
loop:
	for _, v := range names {
		for _, iv := range f {
			if iv.Name == v {
				newList[index] = iv
				//Strip the MetaIgnore since we select the field. Useful for password fields.
				newList[index].Meta &^= MetaIgnore
				index++
				continue loop
			}
		}
	}
	return newList
}

func (f FieldList) Len() int {
	return len(f)
}

//NewList simply represents a list of fields where no copying is needed.
type NewList FieldList

func (f NewList) Get() []Field {
	return f
}

func (f NewList) Len() int {
	return len(f)
}

//Sub allows you to add sub-field structures.
func (f FieldList) Sub(name Predicate, fl Fields) Fields {
	var newArr NewList = make([]Field, len(f))
	//Copy!
	n := copy(newArr, f)
	if n != len(f) {
		panic("fieldList sub: invalid length! something went wrong")
	}

	//linear search but there are not a lot of values. Hash-map feels overkill
	for k, v := range newArr {
		if v.Name == name {
			if fl == nil {
				newArr[k] = Field{
					Name:   v.Name,
					Fields: emptyList,
					Meta:   v.Meta,
				}
				return newArr
			}
			var newField = Field{
				Name:   v.Name,
				Fields: fl,
				Meta:   v.Meta,
			}
			newArr[k] = newField
			break
		}
	}
	return newArr
}

func (f FieldList) Add(fi Field) Fields {
	var newList NewList = make([]Field, len(f)+1)
	copy(newList, f)
	newList[len(f)] = fi
	return newList
}

func (f FieldList) AddName(nam Predicate, sch SchemaList) Fields {
	var newList NewList = make([]Field, len(f)+1)
	copy(newList, f)
	newList[len(f)] = sch[nam]
	return newList
}

//These lists do not need copying as they are never global.
func (f NewList) Sub(name Predicate, fl Fields) Fields {
	//linear search but fast either way.
	for k, v := range f {
		if v.Name == name {
			if fl == nil {
				f[k] = Field{
					Name:   v.Name,
					Fields: emptyList,
					Meta:   v.Meta,
				}
				return f
			}
			f[k].Fields = fl
			break
		}
	}
	return f
}

//facet adds a field of type facet.
func (f NewList) Add(fi Field) Fields {
	return append(f, fi)
}

//PaginationType simply refers to a type of pagination.
type PaginationType string

const (
	CountFirst  PaginationType = "first"
	CountOffset PaginationType = "offset"
	CountAfter  PaginationType = "after"
)

//AggregateType simply refers to all types of aggregation as specified in the query docs.
type AggregateType string

//Types of aggregations.
const (
	Val   AggregateType = ""
	Min   AggregateType = "min"
	Sum   AggregateType = "sum"
	Max   AggregateType = "max"
	Avg   AggregateType = "avg"
	Count AggregateType = "count"
)

type pagination struct {
	Type  PaginationType
	Value int
}

func (c pagination) canApply(mt modifierSource) bool {
	return true
}

func (c pagination) parenthesis() bool {
	return true
}

func (c pagination) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	sb.WriteString(string(c.Type))
	sb.WriteString(tokenColumn)
	sb.WriteString(strconv.Itoa(c.Value))
	return nil
}

func (c pagination) priority() modifierType {
	return modifierPagination
}

//A meta field for schemas.
//This simply defines properties surrounding fields such as language etc.
//This is used in generating the queries.
type FieldMeta uint16

func (f FieldMeta) Lang() bool {
	return f&MetaLang > 0
}

func (f FieldMeta) List() bool {
	return f&MetaList > 0
}

func (f FieldMeta) Reverse() bool {
	return f&MetaReverse > 0
}

func (f FieldMeta) Object() bool {
	return f&MetaObject > 0
}

func (f FieldMeta) Facet() bool {
	return f&MetaFacet > 0
}

func (f FieldMeta) Empty() bool {
	return f&MetaEmpty > 0
}

func (f FieldMeta) Ignore() bool {
	return f&MetaIgnore > 0 || f&MetaFacet > 0
}

const (
	MetaObject FieldMeta = 1 << iota
	MetaList
	MetaLang
	MetaUid
	MetaReverse
	MetaFacet
	MetaEmpty
	MetaIgnore
)

// Field is a recursive data struct which represents a GraphQL query field.
type Field struct {
	Meta   FieldMeta
	Fields Fields
	Name   Predicate
}

//Sub here simply uses fields as Field { fields}.
//That is, you use this if you only want to get a relation.
//Name here does not matter actually.
func (f Field) Sub(name Predicate, fields Fields) Fields {
	f.Fields = fields
	return f
}

func (f Field) Len() int {
	return 1
}

func (f Field) Add(fi Field) Fields {
	var fields = make(NewList, 2)
	fields[0] = f
	fields[1] = fi
	return fields
}

/*
func (f Field) facet(facetName string, alias string) fields {
	return append(NewList{}, f, MakeField(Variable(facetName), 0|MetaFacet))
}*/

func (f Field) Get() []Field {
	return []Field{f}
}

// MakeField constructs a Field of given name and returns the Field.
func MakeField(name Predicate, meta FieldMeta) Field {
	//TODO: better facet support
	var x = Field{Name: name, Meta: meta}
	return x
}

// One may have noticed that there is a public create and a private create.
// The different being the public method checks the validity of the Field structure
// while the private counterpart assumes the validity.
// Returns whether this field is a facet field.
// Parent is in-fact the current field name from the previous level.
func (f *Field) create(q *GeneratedQuery, parent []byte, sb *strings.Builder) error {
	//If a field is an object and has no fields do not use it.
	val, ok := q.modifiers[Predicate(parent)]
	var withGroup, withFacets, withFields bool
	if ok {
		withGroup = len(val.g.m) != 0
		withFacets = val.f.active
		withFields = withGroup || withFacets
	}
	if f.Meta.Object() && ((f.Fields != nil && f.Fields.Len() == 0) || f.Fields == nil) {
		if !withFields {
			return nil
		}
	}
	if f.Meta.Ignore() {
		return nil
	}
	if f.Meta.Lang() {
		sb.WriteString(string(f.Name))
		sb.WriteByte('@')
		sb.WriteString(string(q.language))
		if !q.strictLanguage {
			sb.WriteString(":.")
		}
	} else {
		sb.WriteString(string(f.Name))
	}
	//First part of modifiers, non-field generating.
	if ok {
		//Dont call sort for size 1,2 (common sizes)
		val.m.sort()
		err := val.m.runNormal(q, f.Meta, modifierField, sb)
		if err != nil {
			return err
		}
		if withFacets {
			err = val.f.apply(q, f.Meta, modifierField, sb)
			if err != nil {
				return err
			}
		}
	}
	if withGroup {
		err := val.g.apply(q, 0, modifierField, sb)
		if err != nil {
			return err
		}
	}

	if f.Fields != nil && f.Fields.Len() > 0 {
		if f.Meta.Lang() {
			return errors.New("cannot have language meta and children fields")
		}
		sb.WriteByte('{')
		if !f.Meta.Empty() && f.Fields != nil {
			for i, field := range f.Fields.Get() {
				if len(field.Name) > 0 {
					if i != 0 {
						sb.WriteByte(' ')
					}
					parent = append(parent, field.Name...)
					err := field.create(q, parent, sb)
					if err != nil {
						return err
					}
					parent = parent[:len(parent)-len(field.Name)]
				}
			}
		}
		if ok {
			err := val.m.runVariables(q, f.Meta, modifierField, sb)
			if err != nil {
				return err
			}
		}
		sb.WriteString(" uid}")
	}
	//Always add the uid field. I don't think this will be very expensive in terms of dgraph performance.
	return nil
}
