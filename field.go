package mulbase

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
)

/*
	FieldList is a list of Fields associated with a generated object.
	These are of global type and should never be modified lest
	the state of the entire application should be changed.
	Whenever fields are added to this list they are copied.
	Example usage:

	var NewFields = CharacterFields.Sub("Character.friends", CharacterFields).Sub("Character.enemies", CharacterFields.
					Sub("Character.items", ItemFields)
	This will also ensure fields are copied properly from the global list.
*/

type Fields interface {
	Sub(name Predicate, fields Fields) Fields
	Add(fi Field) Fields
	Facet(facetName string, alias string) Fields
	Get() []Field
	Len() int
}

func Select(sch SchemaList, names ...Predicate) Fields {
	var fields = make(NewList, len(names))
	for k, v := range names {
		fields[k] = sch[v]
	}
	return fields
}

type FieldList []Field

func (f FieldList) Get() []Field {
	return f
}

func (f FieldList) Len() int {
	return len(f)
}

//No need to reallocate. Typing here is very important as it will no longer copy.
type NewList FieldList

func (f NewList) Get() []Field {
	return f
}

func (f NewList) Len() int {
	return len(f)
}

//TODO: Racy reads? There are never writes to a global field-list unless you are doing something wrong!
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

//Facet adds a field of type facet.
func (f FieldList) Facet(facetName string, alias string) Fields {
	var newArr NewList = make([]Field, len(f)+1)
	//Copy!
	n := copy(newArr, f)
	if n != len(f) {
		panic("fieldList sub: invalid length! something went wrong")
	}
	newArr[len(f)] = MakeField(Predicate(facetName), 0|MetaFacet)
	return newArr
}

//These lists do not need copying as they are never global.
func (f NewList) Sub(name Predicate, fl Fields) Fields {
	//linear search but fast either way.
	for k, v := range f {
		if v.Name == name {
			f[k].Fields = fl
			break
		}
	}
	return f
}

func (f NewList) Facet(facetName string, alias string) Fields {
	return append(f, MakeField(Predicate(facetName), 0|MetaFacet))
}

func Sub(name Predicate, fl []Field, sch SchemaList) NewList {
	var newFields = make([]Field, 1)
	newFields[0] = sch[name]
	newFields[0].Fields = NewList(fl)
	return newFields
}

/*func Fields(sch SchemaList, names ...string) NewList{
	var ret = make([]Field, len(names))
	for k,v := range names {
		val := sch[v]
		ret[k] = val
	}
	return ret
}*/
//Facet adds a field of type facet.
func (f NewList) Add(fi Field) Fields {
	return append(f, fi)
}

//CreatePath creates a nested path using paths of the form Pred1/Pred2...
func CreatePath(paths ...Predicate) Predicate {
	var sum = 0
	for _, v := range paths {
		sum += len(v)
	}
	var buf = strings.Builder{}
	buf.Grow(sum + len(paths) - 1)
	for _, v := range paths {
		buf.WriteString(string(v))
		buf.WriteByte('/')
	}
	return Predicate(buf.String())
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
//This simply defines properties surrounding fields such as language etc.
//This is used in generating the queries.
type FieldMeta uint16

func (f FieldMeta) Lang() bool {
	return f&MetaLang > 0
}

func (f FieldMeta) List() bool {
	return f&MetaLang > 0
}

func (f FieldMeta) Reverse() bool {
	return f&MetaReverse > 0
}

func (f FieldMeta) Object() bool {
	return f&MetaObject > 0
}

const (
	MetaObject FieldMeta = 1 << iota
	MetaList
	MetaLang
	MetaUid
	MetaReverse
	MetaFacet
)

// Field is a recursive data struct which represents a GraphQL query field.
type Field struct {
	Meta   FieldMeta
	Name   Predicate
	Fields Fields
	//Type VarType
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

func (f Field) Facet(facetName string, alias string) Fields {
	return append(NewList{}, f, MakeField(Predicate(facetName), 0|MetaFacet))
}

func (f Field) Get() []Field {
	return []Field{f}
}

// MakeField constructs a Field of given name and returns the Field.
func MakeField(name Predicate, meta FieldMeta) Field {
	//TODO: better facet support
	var x = Field{Name: name, Meta: meta}
	return x
}

func (f *Field) writeLanguageTag(sb *bytes.Buffer, l Language) {
	if l != LanguageDefault && l != LanguageNone {
		sb.WriteString("@" + string(l) + ":.")
	}
}

func (f *Field) getName() Predicate {
	if f.Name[0] == '~' {
		return f.Name[1:]
	}
	return f.Name
}

// One may have noticed that there is a public create and a private create.
// The different being the public method checks the validity of the Field structure
// while the private counterpart assumes the validity.
// Returns whether this field is a facet field.
func (f *Field) create(q *GeneratedQuery, parent Predicate, sb *bytes.Buffer) {
	//Default signature of field, do not include.
	if f.Meta.Object() && (f.Fields != nil && f.Fields.Len() == 0) {
		return
	}
	agg, ok := q.FieldAggregate[parent]
	if ok {
		sb.WriteString(agg.Alias)
		sb.WriteString(" : ")
		sb.WriteString(string(agg.Type))
		sb.WriteString(tokenLP)
		sb.WriteString(string(f.Name))
		//if f.SchemaField.Lang {
		//	f.writeLanguageTag(sb, q.Language)
		//}
		sb.WriteString(tokenRP)
	} else {
		sb.WriteString(string(f.Name))
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
	//Should we order on this field?
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
	//Should we filter on this field?
	filter, ok := q.FieldFilters[parent]
	if ok {
		sb.WriteString(tokenSpace)
		filter.create(q, parent, sb)
	}
	var tmp bytes.Buffer
	if f.Fields != nil && f.Fields.Len() > 0 {
		tmp.WriteString(tokenLB)
		for i, field := range f.Fields.Get() {
			if len(field.Name) > 0 {
				if i != 0 && i != f.Fields.Len() {
					tmp.WriteString(tokenSpace)
				}
				field.create(q, parent+"/"+field.Name, &tmp)
			}
		}
		//Always add the uid field. I don't think this will be very expensive.
		tmp.WriteString(tokenSpace)
		tmp.WriteString("uid")
		tmp.WriteString(tokenSpace)
		tmp.WriteString(tokenRB)
	}
	_, _ = io.Copy(sb, &tmp)
}

func (f *Field) check(q *GeneratedQuery) error {
	if f.Name == "" {
		return errors.New("missing name in field")
	}
	if f.Fields != nil {
		for _, v := range f.Fields.Get() {
			if err := v.check(q); err != nil {
				return err
			}
		}
	}
	return nil
}
